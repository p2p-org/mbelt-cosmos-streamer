package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	cosmosTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/prometheus/common/log"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
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
	GetBlock(height int64) *types.Block
	GetBlockRpc(height int64) *types.Block
	GetTx(txHash string) *cosmosTypes.StdTx
	GetTxByHash(txHash string) *ctypes.ResultTx
	GetTxsRpc(height int64) []*ctypes.ResultTx
	Stop()
	ResendBlock(blockHeight uint64)
	ResendTx(txHash string, index uint32)
}

// NodeConnector info
type ClientApi struct {
	wsURL    string
	rpcURL   string
	lcdURL   string
	wsClient *client.HTTP
}

func (nm *ClientApi) Init(cfg *config.Config) error {
	nm.wsURL = "http://" + cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.WebSocketPort)
	nm.rpcURL = cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.RPCPort)
	nm.lcdURL = cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.LCDPort)

	return nil
}

func (nm *ClientApi) Connect() error {
	nm.wsClient = client.NewHTTP(nm.wsURL, "/websocket")
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
	nm.wsClient.Stop()
}

func (nm *ClientApi) GetBlockRpc(height int64) *types.Block {
	block, err := nm.wsClient.Block(&height)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	return block.Block
}

func (nm *ClientApi) GetTxsRpc(height int64) []*ctypes.ResultTx {
	txs, err := nm.wsClient.TxSearch(fmt.Sprintf("tx.height=%d", height), true, 1, 1000)
	if err != nil {
		log.Errorln(err)
		return []*ctypes.ResultTx{}
	}
	return txs.Txs
}

func (nm *ClientApi) GetTxByHash(hash string) *ctypes.ResultTx {
	byteHash, err := hex.DecodeString(hash)
	if err != nil {
		log.Errorln(err)
		return nil
	}
	tx, err := nm.wsClient.Tx(byteHash, true)
	if err != nil {
		log.Errorln(err)
	}
	return tx
}
