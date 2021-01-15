package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/prometheus/common/log"
	client "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
)

const (
	blockQuery string = "tm.event = 'NewBlock'"
	txQuery    string = "tm.event = 'Tx'"
)

type StatusEnum string

const (
	PendingStatus   StatusEnum = "pending"
	ConfirmedStatus StatusEnum = "confirmed"
	RejectedStatus  StatusEnum = "rejected"
	OnForkStatus    StatusEnum = "onfork"
)

type Api interface {
	Init(cfg *config.Config) error
	Connect() error
	SubscribeBlock(ctx context.Context) <-chan ctypes.ResultEvent
	SubscribeTxs(ctx context.Context) <-chan ctypes.ResultEvent
	GetBlockRpc(ctx context.Context, height int64) *types.Block
	GetTx(ctx context.Context, txHash string) *ctypes.ResultTx
	GetTxGrpc(ctx context.Context, hash string) *ctypes.ResultTx
	GetTxsRpc(height int64) []*ctypes.ResultTx
	Stop()
	ResendBlock(blockHeight uint64)
	ResendTx(txHash string, index uint32)
}

// NodeConnector info
type ClientApi struct {
	wsURL      string
	rpcURL     string
	lcdURL     string
	grpcURL    string
	wsClient   *client.HTTP
	grpcClient *grpc.ClientConn
}

func (nm *ClientApi) Init(cfg *config.Config) error {
	nm.wsURL = "http://" + cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.WebSocketPort)
	nm.rpcURL = cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.RPCPort)
	nm.lcdURL = cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.LCDPort)
	nm.grpcURL = cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.GRPCPort)

	return nil
}

func (nm *ClientApi) ConnectWs() error {
	nm.wsClient, _ = client.New(nm.wsURL, "/websocket")
	for {
		err := nm.wsClient.Start()
		if err != nil {
			log.Errorln(nm.wsURL, "err:", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}

	return nil
}

func (nm *ClientApi) ConnectGrpc() error {
	var err error
	nm.grpcClient, err = grpc.Dial(
		nm.grpcURL,
		grpc.WithInsecure(),
	)
	return err
}

func (nm *ClientApi) SubscribeBlock(ctx context.Context) <-chan ctypes.ResultEvent {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	blocks, err := nm.wsClient.Subscribe(ctx, "test-client", blockQuery)
	if err != nil {
		log.Errorln(err)
	}
	return blocks
}

func (nm *ClientApi) SubscribeTxs(ctx context.Context) <-chan ctypes.ResultEvent {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	txs, err := nm.wsClient.Subscribe(ctx, "test-client", txQuery)
	if err != nil {
		log.Errorln(err)
	}
	return txs
}

func (nm *ClientApi) Stop() {
	if err := nm.wsClient.Stop(); err != nil {
		log.Warnln(err)
	}
}

func (nm *ClientApi) GetBlockRpc(ctx context.Context, height int64) *types.Block {
	block, err := nm.wsClient.Block(ctx, &height)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	return block.Block
}

func (nm *ClientApi) GetTxsRpc(ctx context.Context, height int64) []*ctypes.ResultTx {
	page := 1
	peerPage := 1000
	str := fmt.Sprintf("tx.height=%d", height)
	txs, err := nm.wsClient.TxSearch(ctx, str, true, &page, &peerPage, "asc")
	if err != nil {
		log.Errorln(err)
		return []*ctypes.ResultTx{}
	}
	return txs.Txs
}

func (nm *ClientApi) GetTx(ctx context.Context, hash string) *ctypes.ResultTx {
	byteHash, err := hex.DecodeString(hash)
	if err != nil {
		log.Errorln(err)
		return nil
	}
	tx, err := nm.wsClient.Tx(ctx, byteHash, true)
	if err != nil {
		log.Errorln(err)
		return nil
	}
	return tx
}

func (nm *ClientApi) GetTxGrpc(ctx context.Context, hash string) *tx.GetTxResponse {
	txClient := tx.NewServiceClient(nm.grpcClient)
	newTx, err := txClient.GetTx(ctx, &tx.GetTxRequest{Hash: hash})
	if err != nil {
		log.Errorln(err)
	}
	return newTx
}

//
// func (nm *ClientApi) ResendBlock(blockHeight uint64) {
// 	url := "http://" + nm.lcdURL + "/blocks/" + strconv.FormatUint(blockHeight, 10)
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		log.Errorf("ResendBlock got responce error: %v", err)
// 	}
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Errorf("ResendBlock problem in decode responce to byte array. error: %v", err)
// 	}
// 	var block ctypes.ResultBlock
// 	err = cdc.UnmarshalJSON(body, &block)
// 	if err != nil {
// 		log.Errorf("ResendBlock problem in unmarshal to resultBlock. error: %v", err)
// 	}
// 	blockToProduce := Block{
// 		Hash:          fmt.Sprintf("%X", block.BlockMeta.BlockID.Hash),
// 		ChainID:       block.Block.Header.ChainID,
// 		Height:        uint64(block.Block.Header.Height),
// 		Time:          block.Block.Header.Time,
// 		NumTxs:        uint64(block.Block.Header.NumTxs),
// 		TotalTxs:      uint64(block.Block.Header.TotalTxs),
// 		LastBlockHash: fmt.Sprintf("%X", block.Block.Header.LastBlockID.Hash),
// 		Validator:     fmt.Sprintf("%X", block.Block.Header.ValidatorsHash),
// 		Status:        PendingStatus,
// 	}
// 	for i, temp_tx := range block.Block.Data.Txs {
// 		var txResult cosmosTypes.StdTx
// 		err := cdc.UnmarshalBinaryLengthPrefixed(temp_tx, &txResult)
// 		if err != nil {
// 			log.Errorf("error parse Tx err: %v,  Txresult: %v", err, txResult)
// 		}
// 		// fee, err := cdc.MarshalJSON(txResult.Fee)
// 		// if err != nil {
// 		// 	log.Errorf("error on marshal fee to json err %v data %v", err, txResult.Fee)
// 		// }
// 		// signatures, err := cdc.MarshalJSON(txResult.Signatures)
// 		// if err != nil {
// 		// 	log.Errorf("error on marshal signature to json err %v data %v", err, txResult.Fee)
// 		// }
// 		messages := make([]Message, 0)
//
// 		for _, msg := range txResult.GetMsgs() {
// 			message, err := cdc.MarshalJSON(msg)
// 			if err != nil {
// 				log.Errorf("err on marshal to JSON  err : %v", err)
// 			}
// 			var tempMessage struct {
// 				MsgType string                 `json:"type"`
// 				Msg     map[string]interface{} `json:"value"`
// 			}
//
// 			err = json.Unmarshal(message, &tempMessage)
// 			if err != nil {
// 				log.Errorf("err on Unmarshal from JSON  err : %v, data:%v\n\n", err, string(message))
// 			}
// 			msgValue, err := json.Marshal(tempMessage.Msg)
// 			if err != nil {
// 				log.Errorf("err on Marshal to JSON messageValue err : %v, data:%v\n\n", err, string(message))
// 			}
// 			resultMessage := Message{
// 				MsgType: tempMessage.MsgType,
// 				Msg:     string(msgValue),
// 			}
// 			messages = append(messages, resultMessage)
// 		}
// 		// blockToProduce.Txs = append(blockToProduce.Txs, BlockTransaction{
// 		// 	TxHash:     fmt.Sprintf("%X", temp_tx.Hash()),
// 		// 	Messages:   messages,
// 		// 	Fee:        string(fee),
// 		// 	Signatures: string(signatures),
// 		// 	Memo:       txResult.GetMemo(),
// 		// 	Status:     PendingStatus,
// 		// })
// 		go nm.ResendTx(fmt.Sprintf("%X", temp_tx.Hash()), uint32(i))
// 	}
//
// 	nm.BlockChan <- blockToProduce
// }
//
// func (nm *ClientApi) ResendTx(txHash string, index uint32) {
// 	url := "http://" + nm.lcdURL + "/txs/" + txHash
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		log.Errorf("Error: %v", err)
// 	}
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Errorf("Error: %v", err)
// 	}
// 	var tx sdk.TxResponse
// 	err = cdc.UnmarshalJSON(body, &tx)
// 	if err != nil {
// 		log.Errorf("error on unmarshal txResult from json err %v data %v\n", err, body)
// 	}
//
// 	txResult := tx.Tx.(cosmosTypes.StdTx)
// 	signatures, err := json.Marshal(txResult.Signatures)
// 	if err != nil {
// 		log.Errorf("error on marshal signature to json err %v data %v\n", err, txResult.Fee)
// 	}
//
// 	var messages []Message
//
// 	var good bool = true
// 	for _, log_info := range tx.Logs {
// 		if !log_info.Success {
// 			good = false
// 		}
// 		var tempMessage TempMessage
//
// 		msg, err := cdc.MarshalJSON(txResult.Msgs[log_info.MsgIndex])
// 		if err != nil {
// 			log.Errorf("err on marshal msg to JSON  err : %v\n\n", err)
// 		}
// 		err = json.Unmarshal(msg, &tempMessage)
// 		if err != nil {
// 			log.Errorf("err on Unmarshal from JSON  err : %v, data:%v\n\n", err, string(msg))
// 		}
// 		msgValue, err := json.Marshal(tempMessage.Msg)
// 		if err != nil {
// 			log.Errorf("err on Marshal to JSON messageValue err : %v, data:%v\n\n", err, string(msg))
// 		}
// 		events, err := json.Marshal(log_info.Events)
// 		if err != nil {
// 			log.Errorf("err on marshal event to JSON  err : %v\n\n", err)
// 		}
// 		message := Message{
// 			MsgType: tempMessage.MsgType,
// 			Msg:     string(msgValue),
// 			Events:  string(events),
// 			Log:     log_info.Log,
// 		}
// 		messages = append(messages, message)
// 	}
//
// 	fee := Fee{
// 		GasWanted: fmt.Sprintf("%d", tx.GasWanted),
// 		GasUsed:   fmt.Sprintf("%d", tx.GasUsed),
// 		Amount:    txResult.Fee.Amount.String(),
// 	}
// 	feeString, err := json.Marshal(fee)
// 	if err != nil {
// 		log.Errorf("err on Marshal fee to JSON, err : %v, data:%v\n\n", err, fee)
// 	}
// 	txToProduce := BlockTransaction{
// 		TxHash:     tx.TxHash,
// 		Messages:   messages,
// 		Fee:        string(feeString),
// 		Signatures: string(signatures),
// 		Memo:       txResult.GetMemo(),
// 	}
// 	if good && len(tx.Logs) == len(messages) {
// 		txToProduce.Status = ConfirmedStatus
// 	}
//
// 	nm.TxChan <- TransactionWithBlockInfo{
// 		BlockNum: tx.Height,
// 		TxNum:    index,
// 		Tx:       txToProduce,
// 	}
//
// }
