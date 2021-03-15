package client

import (
	"context"
	"fmt"
	"strconv"
	"time"

	clientContext "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types/tx"
	client2 "github.com/cosmos/cosmos-sdk/x/auth/client"
	app "github.com/cosmos/gaia/v4/app"
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

var cdc, _ = app.MakeCodecs()

// NodeConnector info
type ClientApi struct {
	wsURL         string
	rpcURL        string
	lcdURL        string
	grpcURL       string
	wsClient      *client.HTTP
	grpcClient    *grpc.ClientConn
	contextClient clientContext.Context
}

func (nm *ClientApi) Init(cfg *config.Config) error {
	nm.wsURL = "http://" + cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.WebSocketPort)
	nm.rpcURL = cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.RPCPort)
	nm.lcdURL = cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.LCDPort)
	nm.grpcURL = cfg.Node.Host + ":" + strconv.Itoa(cfg.Node.GRPCPort)

	if err := nm.connectWs(); err != nil {
		return err
	}

	if err := nm.connectGrpc(); err != nil {
		return err
	}

	cctx := clientContext.Context{}
	cctx = cctx.WithClient(nm.wsClient).WithTxConfig(app.MakeEncodingConfig().TxConfig)
	nm.contextClient = cctx

	return nil
}

func (nm *ClientApi) connectWs() error {
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

func (nm *ClientApi) connectGrpc() error {
	var err error
	nm.grpcClient, err = grpc.Dial(
		nm.grpcURL,
		grpc.WithInsecure(),
	)
	return err
}

func (nm *ClientApi) SubscribeBlock(ctx context.Context) <-chan ctypes.ResultEvent {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	blocks, err := nm.wsClient.Subscribe(ctx, "test-client", blockQuery)
	if err != nil {
		log.Errorln(err)
	}
	return blocks
}

func (nm *ClientApi) SubscribeTxs(ctx context.Context) <-chan ctypes.ResultEvent {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
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

func (nm *ClientApi) GetTx(ctx context.Context, hash string) *sdk.GetTxResponse {
	queryTx, err := client2.QueryTx(nm.contextClient, hash)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	var newTx sdk.Tx
	if err := cdc.UnmarshalBinaryBare(queryTx.Tx.Value, &newTx); err != nil {
		log.Errorln(err)
		return nil
	}

	return &sdk.GetTxResponse{
		Tx:         &newTx,
		TxResponse: queryTx,
	}
}

func (nm *ClientApi) GetTxGrpc(ctx context.Context, hash string) *sdk.GetTxResponse {
	txClient := sdk.NewServiceClient(nm.grpcClient)

	newTx, err := txClient.GetTx(ctx, &sdk.GetTxRequest{Hash: hash})
	if err != nil {
		log.Errorln(err)
		return nm.GetTx(ctx, hash)
	}
	return newTx
}
