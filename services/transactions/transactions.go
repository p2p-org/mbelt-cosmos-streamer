package transactions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/tx"
	app "github.com/cosmos/gaia/v3/app"
	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/utils"
	"github.com/prometheus/common/log"
)

var cdc, _ = app.MakeCodecs()

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

func Init(config *config.Config, ds *datastore.KafkaDatastore, pgDs *pg.PgDatastore) (*Service, error) {
	return &Service{
		config: config,
		ds:     ds,
	}, nil
}

func (s *Service) Push(txData *tx.GetTxResponse) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Fatalln("[TransactionsService][Recover]", "Throw panic", r)
	// 	}
	// }()
	txPush := map[string]interface{}{}
	txResult := s.serialize(txData)
	msgs := txResult["messages_for_push"]
	delete(txResult, "messages_for_push")

	txPush[fmt.Sprintf("%X", txData.TxResponse.TxHash)] = txResult

	msgsPush := map[string]interface{}{}

	for _, msg := range msgs.([]map[string]interface{}) {
		key := fmt.Sprintf("%s-%d", txData.TxResponse.TxHash, msg["msg_index"])
		msgsPush[key] = msg
	}

	if err := s.ds.Push(datastore.TopicTransactions, txPush); err != nil {
		log.Warnln(err)
	}
	if err := s.ds.Push(datastore.TopicMessages, msgsPush); err != nil {
		log.Warnln(err)
	}
}

func (s *Service) serialize(txData *tx.GetTxResponse) map[string]interface{} {
	signatures, err := json.Marshal(txData.Tx.Signatures)
	if err != nil {
		log.Errorf("error on marshal signature to json err %v data \n", err)
	}

	messages := []map[string]interface{}{}
	messagesForPush := []map[string]interface{}{}
	for _, logInfo := range txData.TxResponse.Logs {
		if &logInfo == nil {
			log.Errorln("struct is nil logInfo")
		}
		events := make([]map[string]interface{}, len(logInfo.Events))
		for _, event := range logInfo.Events {
			if &event != nil {
				events = append(events, map[string]interface{}{
					"type":       event.Type,
					"attributes": event.Attributes,
				})
			}
		}

		msg := txData.Tx.Body.Messages[logInfo.MsgIndex]
		msgValue := cdc.MustMarshalJSON(msg)
		messageForPush := map[string]interface{}{
			"block_height":  txData.TxResponse.Height,
			"tx_hash":       txData.TxResponse.TxHash,
			"tx_index":      0, // TODO add value
			"msg_index":     logInfo.MsgIndex,
			"msg_type":      msg.TypeUrl,
			"msg_info":      string(msgValue),
			"logs":          logInfo.Log,
			"events":        events,
			"external_info": utils.ToVarcharArray([]string{}),
		}
		messagesForPush = append(messagesForPush, messageForPush)

		message := map[string]interface{}{
			"type":   msg.TypeUrl,
			"value":  string(msgValue),
			"events": events,
			"log":    logInfo.Log,
		}
		messages = append(messages, message)
	}

	result := map[string]interface{}{
		"tx_hash":        txData.TxResponse.TxHash,
		"chain_id":       s.config.ChainID,
		"block_height":   txData.TxResponse.Height,
		"time":           txData.TxResponse.Timestamp,
		"tx_index":       0, // TODO add tx_index txData.TxResponse.Index,
		"count_messages": len(txData.Tx.Body.Messages),
		"logs":           txData.TxResponse.RawLog,
		"events":         txData.TxResponse.Logs.String(),
		"msgs":           messages,
		"fee": map[string]interface{}{
			"gas_wanted": fmt.Sprintf("%d", txData.TxResponse.GasWanted),
			"gas_used":   fmt.Sprintf("%d", txData.TxResponse.GasUsed),
			"gas_amount": txData.Tx.AuthInfo.Fee.Amount.String(),
		},
		"signatures":        string(signatures),
		"memo":              strconv.Quote(strings.ReplaceAll(txData.Tx.Body.Memo, "'", "`")),
		"status":            client.PendingStatus,
		"external_info":     utils.ToVarcharArray([]string{}),
		"messages_for_push": messagesForPush,
	}

	return result
}
