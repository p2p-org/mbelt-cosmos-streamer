package messages

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/common/log"

	cosmosTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gaia/app"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore"
	"github.com/tendermint/tendermint/types"
)

type StatusEnum string

const (
	PendingStatus   StatusEnum = "pending"
	ConfirmedStatus StatusEnum = "confirmed"
	RejectedStatus  StatusEnum = "rejected"
	OnForkStatus    StatusEnum = "onfork"
)

type Service struct {
	config *config.Config
	ds     *datastore.KafkaDatastore
}

var cdc = app.MakeCodec()

func Init(config *config.Config, ds *datastore.KafkaDatastore) (*Service, error) {
	return &Service{
		config: config,
		ds:     ds,
	}, nil
}

func (s *Service) Push(tx *types.EventDataTx) {
	// Empty messages has panic
	defer func() {
		if r := recover(); r != nil {
			log.Fatalln("[TransactionsService][Recover]", "Throw panic", r)
		}
	}()

	m := map[string]interface{}{}
	m[fmt.Sprintf("%X", tx.Tx.Hash())] = s.serialize(tx)

	s.ds.Push(datastore.TopicTransactions, m)
}

func (s *Service) serialize(tx *types.EventDataTx) map[string]interface{} {
	var txResult cosmosTypes.StdTx
	err := cdc.UnmarshalBinaryLengthPrefixed([]byte(tx.Tx), &txResult)
	// txResult, err := cosmosTypes.ParseTx(cdc, []byte(tx.Tx))
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
	err = json.Unmarshal([]byte(tx.Result.Log), &logs)

	result := map[string]interface{}{
		"tx_hash":      fmt.Sprintf("%X", tx.Tx.Hash()),
		"chain_id":     "cosmoshub-3",
		"block_height": "",
		"block_hash":   "",
		"time":         "",
		"tx_index":     "",
		"logs":         "",
		"events":       "",
		"msgs":         "",
		"fee": map[string]interface{}{
			"gas_wanted": fmt.Sprintf("%d", tx.Result.GasWanted),
			"gas_used":   fmt.Sprintf("%d", tx.Result.GasUsed),
			"gas_amount": txResult.Fee.Amount.String(),
		},
		"signatures":    string(signatures),
		"memo":          "",
		"status":        ConfirmedStatus,
		"external_info": "",
	}
	return result
}
