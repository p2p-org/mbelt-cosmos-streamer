package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"google.golang.org/grpc"

	app "github.com/cosmos/gaia/v3/app"

	"github.com/jinzhu/configor"
	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/prometheus/common/log"
)

var cfg config.Config

var cdc, cdcAmino = app.MakeCodecs()

func main() {

	if err := configor.Load(&cfg); err != nil {
		log.Fatal(err)
	}
	syncCtx, syncCancel := context.WithCancel(context.Background())

	api := &client.ClientApi{}
	// api := &cache.CacheApi{Api: &client.ClientApi{}}
	if err := api.Init(&cfg); err != nil {
		log.Fatalln(err)
	}

	if err := api.ConnectGrpc(); err != nil {
		log.Fatalln(err)
	}

	go func() {
		var gracefulStop = make(chan os.Signal)
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		signal.Notify(gracefulStop, syscall.SIGHUP)

		sig := <-gracefulStop
		log.Infof("Caught sig: %+v", sig)
		log.Infoln("Wait for graceful shutdown to finish.")

		syncCancel()
		api.Stop()
	}()

	for _, txNew := range api.GetTxsRpc(syncCtx, 369) {
		var result tx.Tx
		log.Infoln(string(txNew.Tx))
		// val.ClientCtx.JSONMarshaler.
		if err := cdcAmino.UnmarshalJSON(txNew.Tx, &result); err != nil {
			log.Fatalln(err)
		}
		log.Infoln(result)
		// var txDd cosmosTypes.Tx
		// // if err :=  df.UnmarshalBinaryLengthPrefixed(tx.Tx, &txDd); err != nil {
		// // 	log.Errorln(err)
		// // }
		// if err :=  cdc.UnmarshalJSON(tx.Tx, &txDd); err != nil {
		// 	log.Errorln(err)
		// }
		// // if err := txDd.Unmarshal(tx.Tx); err != nil {
		// // 	log.Errorln(err)
		// // }
		// log.Infoln(txDd)
		// var newTx cosmosTypes.Tx
		// if err := json.Unmarshal(tx.Tx, &newTx); err != nil {
		// 	log.Errorln(err)
		// }
		// log.Infoln(newTx)

	}

	// GetTx("64DDF005BA6110BA42DAAC8870F1354821E28EF41C4D5BD4A962358309FFD2F1")
	// grpc2()
	// TODO memo signatures events msgs logs
	<-syncCtx.Done()
}

//
// func GetTx(txHash string) {
// 	url := "http://159.69.196.16:1317/txs/" + txHash
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		log.Errorf("Error: %v", err)
// 		// return nil
// 	}
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Errorf("Error: %v", err)
// 		// return nil
// 	}
// 	// log.Infoln(string(body))
// 	var tx sdk.TxResponse
//
// 	if err != nil {
// 		log.Errorf("error on unmarshal txResult from json err %v \n", err)
// 		// return nil
// 	}
// 	//
// 	// // newTx := tx.Tx.Value.(cosmosTypes.Tx)
// 	//
// 	log.Infoln(tx)
// }
//
func grpc2(args ...string) {
	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		"159.69.196.16:9090", // Or your gRPC server address.
		grpc.WithInsecure(),  // The SDK doesn't support any transport security mechanism.
	)
	if err != nil {
		log.Errorln(err)
		return
	}
	defer grpcConn.Close()
	log.Infoln(grpcConn.GetState())
	// var reply interface{}
	// log.Infoln(grpcConn.Invoke(context.TODO(), "POST",tx.GetTxRequest{}, reply))
	// Broadcast the tx via gRPC. We create a new client for the Protobuf Tx
	// service.
	txClient := tx.NewServiceClient(grpcConn)
	txClient.GetTx()
	res, err := txClient.GetTxsEvent(context.TODO(), &tx.GetTxsEventRequest{Events: []string{"tm.event='Tx'"}})
	if err != nil {
		log.Errorln(err)
	}

	for _, df := range res.TxResponses {
		log.Infoln(df.TxHash)
		// log.Infoln(df.Body.Memo)
		// log.Infoln(df.AuthInfo.Fee.Amount.String())
	}
	log.Infoln(len(res.TxResponses)) // Should be `0` if the tx is successful
}
