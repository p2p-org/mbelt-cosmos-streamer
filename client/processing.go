package client

import (
	"encoding/json"
	"fmt"

	cosmosTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/prometheus/common/log"
	"github.com/tendermint/tendermint/types"
)

func processingBlock(block types.EventDataNewBlock, chanBlock chan<- Block) {
	blockToProduce := Block{
		Hash:          block.Block.Hash().String(),
		ChainID:       block.Block.Header.ChainID,
		Height:        uint64(block.Block.Header.Height),
		Time:          block.Block.Header.Time,
		NumTxs:        uint64(block.Block.Header.NumTxs),
		TotalTxs:      uint64(block.Block.Header.TotalTxs),
		LastBlockHash: block.Block.Header.LastBlockID.Hash.String(),
		Validators:    block.Block.Header.ValidatorsHash.String(),
		Status:        "pending",
	}
	for _, tx := range block.Block.Data.Txs {
		var txResult cosmosTypes.StdTx
		err := cdc.UnmarshalBinaryLengthPrefixed(tx, &txResult)
		if err != nil {
			log.Errorf("error parse Tx err: %v,  Txresult: %v", err, txResult)
		}

		fee, err := cdc.MarshalJSON(txResult.Fee)
		if err != nil {
			log.Errorf("error on marshal fee to json err %v data %v", err, txResult.Fee)
		}

		signatures, err := cdc.MarshalJSON(txResult.Signatures)
		if err != nil {
			log.Errorf("error on marshal signature to json err %v data %v", err, txResult.Fee)
		}
		messages := make([]Message, 0)
		for _, msg := range txResult.GetMsgs() {
			message, err := cdc.MarshalJSON(msg)
			if err != nil {
				log.Errorf("err on marshal to JSON  err : %v", err)
			}
			var tempMessage TempMessage

			err = json.Unmarshal(message, &tempMessage)
			if err != nil {
				log.Errorf("err on Unmarshal from JSON  err : %v, data:%v\n\n", err, string(message))
			}
			msgValue, err := json.Marshal(tempMessage.Msg)
			if err != nil {
				log.Errorf("err on Marshal to JSON messageValue err : %v, data:%v\n\n", err, string(message))
			}
			resultMessage := Message{
				MsgType: tempMessage.MsgType,
				Msg:     string(msgValue),
			}
			messages = append(messages, resultMessage)
		}

		blockToProduce.Txs = append(blockToProduce.Txs, BlockTransaction{
			TxHash:     fmt.Sprintf("%X", tx.Hash()),
			Messages:   messages,
			Fee:        string(fee),
			Signatures: string(signatures),
			Memo:       txResult.GetMemo(),
			Status:     PendingStatus,
		})
	}
	chanBlock <- blockToProduce
}

func processingTxWithEvents(tx types.EventDataTx, chanTxs chan<- TransactionWithBlockInfo) {
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

	var messages []Message

	var logs []struct {
		MsgIndex uint64      `json:"msg_index"`
		Success  bool        `json:"success"`
		Log      string      `json:"log"`
		Events   interface{} `json:"events"`
	}
	var good bool = true
	err = json.Unmarshal([]byte(tx.Result.Log), &logs)
	if err != nil {
		log.Errorf("error on marshal logs to json err %v data %v\n", err, tx.Result.Log)
		good = false
	} else {
		for _, log_info := range logs {
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
			message := Message{
				MsgType: tempMessage.MsgType,
				Msg:     string(msgValue),
				Events:  string(events),
				Log:     log_info.Log,
			}
			messages = append(messages, message)
		}

	}
	fee := Fee{
		GasWanted: fmt.Sprintf("%d", tx.Result.GasWanted),
		GasUsed:   fmt.Sprintf("%d", tx.Result.GasUsed),
		Amount:    txResult.Fee.Amount.String(),
	}
	feeString, err := json.Marshal(fee)
	if err != nil {
		log.Errorf("err on Marshal fee to JSON, err : %v, data:%v\n\n", err, fee)
	}
	txToProduce := BlockTransaction{
		TxHash:     fmt.Sprintf("%X", tx.Tx.Hash()),
		Messages:   messages,
		Fee:        string(feeString),
		Signatures: string(signatures),
		Memo:       txResult.GetMemo(),
	}
	if good && len(logs) == len(messages) {
		txToProduce.Status = ConfirmedStatus
	}

	chanTxs <- TransactionWithBlockInfo{
		BlockNum: tx.Height,
		TxNum:    tx.Index,
		Tx:       txToProduce,
	}

}
