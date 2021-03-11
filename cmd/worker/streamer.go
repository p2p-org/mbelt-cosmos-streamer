package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	app "github.com/cosmos/gaia/v4/app"
	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/services"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
)

var cdc, _ = app.MakeCodecs()

type Streamer struct {
	Cmd *cobra.Command
}

func (s *Streamer) Init(cfg *config.Config) {
	s.Cmd = &cobra.Command{
		Use:   "streamer",
		Short: "A streamer of cosmos's entities to PostgreSQL DB through Kafka",
		Long: `This app synchronizes with current cosmos state and keeps in sync by subscribing on it's updates.
Entities (blocks, transactions and messages) are being pushed to Kafka. There are also sinks that get
those entities from Kafka streams and push them in PostgreSQL DB.'`,
		Run: func(cmd *cobra.Command, args []string) {
			s.Start(cfg)
		},
	}
}

func (s *Streamer) Start(config *config.Config) {
	exitCode := 0
	defer os.Exit(exitCode)

	err := services.InitServices(config)
	if err != nil {
		log.Infoln("[App][Debug]", "Cannot init services:", err)
		exitCode = 1
		return
	}
	syncCtx, syncCancel := context.WithCancel(context.Background())

	api := &client.ClientApi{}
	// api := &cache.CacheApi{Api: &client.ClientApi{}}
	if err := api.Init(config); err != nil {
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

	go func() {
		for block := range api.SubscribeBlock(syncCtx) {
			log.Infoln("Height:", block.Data.(types.EventDataNewBlock).Block.Header.Height)
			newBlock := block.Data.(types.EventDataNewBlock).Block
			go services.App().BlocksService().Push(newBlock)
		}
	}()

	go func() {
		for subTx := range api.SubscribeTxs(syncCtx) {
			txReform := subTx.Data.(types.EventDataTx)

			ddd := types.Tx(txReform.Tx)
			log.Infoln("tx new -> ", fmt.Sprintf("%d - %X", subTx.Data.(types.EventDataTx).Height, ddd.Hash()))
			hash := fmt.Sprintf("%X", ddd.Hash())
			txResponse := api.GetTx(syncCtx, hash)

			go services.App().TransactionsService().Push(txResponse)
		}
	}()

	<-syncCtx.Done()
	log.Infoln("mbelt-cosmos-streamer gracefully stopped")
}
