package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/services"
	"github.com/p2p-org/mbelt-cosmos-streamer/watcher"
	"github.com/prometheus/common/log"
	"github.com/tendermint/tendermint/types"
)

func StartWatcher(config *config.Config) {
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
	w := &watcher.Watcher{}
	if err := api.Init(config); err != nil {
		log.Fatalln(err)
	}

	if err := w.Init(config); err != nil {
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

	go w.ListenDB()
	for i := 0; i < 1; i++ {
		go func() {
			for {
				select {
				case height := <-w.Subscribe():
					block := api.GetBlockRpc(height)
					if block == nil {
						continue
					}
					log.Infoln("new block -> ", block.Height)
					services.App().BlocksService().Push(block)
					txs := api.GetTxsRpc(block.Height)
					for _, tx := range txs {
						txResult := types.TxResult{
							Height: tx.Height,
							Index:  tx.Index,
							Tx:     tx.Tx,
							Result: tx.TxResult,
						}
						log.Infoln("new tx -> ", txResult.Height)
						services.App().TransactionsService().Push(&txResult)
					}
				}
			}
		}()
	}

	<-syncCtx.Done()
	log.Infoln("mbelt-cosmos-watcher gracefully stopped")
}
