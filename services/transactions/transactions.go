package transactions

import (
	"encoding/json"
	"fmt"

	cosmosTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gaia/app"
	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/utils"
	"github.com/prometheus/common/log"
	"github.com/tendermint/tendermint/types"
)

type TempMessage struct {
	MsgType string                 `json:"type"`
	Msg     map[string]interface{} `json:"value"`
}
type Message struct {
	MsgType      string `json:"type"`
	Msg          string `json:"value"`
	Events       string `json:"events"`
	ExternalInfo string `json:"external_info"`
	Log          string `json:"log"`
}

type Service struct {
	config *config.Config
	ds     *datastore.KafkaDatastore
}

var cdc = app.MakeCodec()

func Init(config *config.Config, ds *datastore.KafkaDatastore, pgDs *pg.PgDatastore) (*Service, error) {
	return &Service{
		config: config,
		ds:     ds,
	}, nil
}

func (s *Service) Push(tx *types.TxResult) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalln("[TransactionsService][Recover]", "Throw panic", r)
		}
	}()
	txPush := map[string]interface{}{}
	txResult := s.serialize(tx)
	txPush[fmt.Sprintf("%X", tx.Tx.Hash())] = txResult

	s.ds.Push(datastore.TopicTransactions, txPush)
}

func (s *Service) serialize(tx *types.TxResult) map[string]interface{} {
	var txResult cosmosTypes.StdTx
	err := cdc.UnmarshalBinaryLengthPrefixed([]byte(tx.Tx), &txResult)

	if err != nil {
		log.Errorf("error parse Tx err: %v,  Txresult: %v", err, txResult)
	}

	signatures, err := json.Marshal(txResult.Signatures)
	if err != nil {
		log.Errorf("error on marshal signature to json err %v data %v\n", err, txResult.Fee)
	}

	var logs []struct {
		MsgIndex uint64      `json:"msg_index"`
		Success  bool        `json:"success"`
		Log      string      `json:"log"`
		Events   interface{} `json:"events"`
	}
	var messages []map[string]interface{}
	var good bool = true

	err = json.Unmarshal([]byte(tx.Result.Log), &logs)
	if err != nil {
		log.Errorf("error on marshal logs to json err %v data %v\n", err, tx.Result.Log)
		good = false
	} else {
		for _, log_info := range logs {
			if &log_info == nil {
				log.Errorln("struct is nil logInfo")
			}
			if !log_info.Success {
				good = false
			}
			var tempMessage TempMessage

			msg, err := cdc.MarshalJSON(txResult.Msgs[log_info.MsgIndex])
			if err != nil {
				log.Errorf("err on marshal msg to JSON  err : %v\n\n", err)
			}
			err = json.Unmarshal(msg, &tempMessage)
			if err != nil {
				log.Errorf("err on Unmarshal from JSON  err : %v, data:%v\n\n", err, string(msg))
			}
			msgValue, err := json.Marshal(tempMessage.Msg)
			if err != nil {
				log.Errorf("err on Marshal to JSON messageValue err : %v, data:%v\n\n", err, string(msg))
			}
			events, err := json.Marshal(log_info.Events)
			if err != nil {
				log.Errorf("err on marshal event to JSON  err : %v\n\n", err)
			}

			message := map[string]interface{}{
				"type":   tempMessage.MsgType,
				"value":  string(msgValue),
				"events": string(events),
				"log":    log_info.Log,
			}
			messages = append(messages, message)
		}
	}

	result := map[string]interface{}{
		"tx_hash":      fmt.Sprintf("%X", tx.Tx.Hash()),
		"chain_id":     s.config.ChainID,
		"block_height": tx.Height,
		"tx_index":     tx.Index,
		"logs":         logs,
		"events":       utils.ToVarcharArray([]string{}),
		"msgs":         messages,
		"fee": map[string]interface{}{
			"gas_wanted": fmt.Sprintf("%d", tx.Result.GasWanted),
			"gas_used":   fmt.Sprintf("%d", tx.Result.GasUsed),
			"gas_amount": txResult.Fee.Amount.String(),
		},
		"signatures":    string(signatures),
		"memo":          txResult.GetMemo(),
		"status":        client.PendingStatus,
		"external_info": utils.ToVarcharArray([]string{}),
	}

	if good && len(logs) == len(messages) {
		result["status"] = client.ConfirmedStatus
	}

	return result
}
