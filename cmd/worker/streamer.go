package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/services"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
)

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

	if err = api.Connect(); err != nil {
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
		for tx := range api.SubscribeTxs(syncCtx) {
			log.Infoln("tx new -> ", fmt.Sprintf("%X %d", tx.Data.(types.EventDataTx).Tx.Hash(), tx.Data.(types.EventDataTx).Height))
			newTx := tx.Data.(types.EventDataTx).TxResult
			go services.App().TransactionsService().Push(&newTx)
		}
	}()

	<-syncCtx.Done()
	log.Infoln("mbelt-cosmos-streamer gracefully stopped")
}
