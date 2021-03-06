package worker

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/services"
	"github.com/p2p-org/mbelt-cosmos-streamer/watcher"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Watcher struct {
	Worker int
	Cmd    *cobra.Command
}

func (w *Watcher) Init(cfg *config.Config) {
	w.Cmd = &cobra.Command{
		Use:   "watcher",
		Short: "A watcher of cosmos's entities to PostgreSQL DB through Kafka",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			w.Start(cfg)
		},
	}
	w.Cmd.PersistentFlags().IntVar(&w.Worker, "worker", 3,
		"How many workers to run for processing")

	viper.SetDefault("worker", 3)
}

func (w *Watcher) Start(config *config.Config) {
	exitCode := 0
	defer os.Exit(exitCode)

	if config.Watcher.Worker != -1 {
		w.Worker = config.Watcher.Worker
	}
	log.Infoln("Start watcher with worker: ", w.Worker)

	err := services.InitServices(config)
	if err != nil {
		log.Infoln("[App][Debug]", "Cannot init services:", err)
		exitCode = 1
		return
	}
	syncCtx, syncCancel := context.WithCancel(context.Background())

	api := &client.ClientApi{}
	watcherDB := &watcher.Watcher{}
	if err := api.Init(config); err != nil {
		log.Fatalln(err)
	}

	if err := watcherDB.Init(config); err != nil {
		log.Fatalln(err)
	}

	wg := &sync.WaitGroup{}
	go func() {
		var gracefulStop = make(chan os.Signal)
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		signal.Notify(gracefulStop, syscall.SIGHUP)

		sig := <-gracefulStop
		log.Infof("Caught sig: %+v", sig)
		log.Infoln("Wait for graceful shutdown to finish.")

		syncCancel()
		wg.Wait()
		api.Stop()
	}()

	if config.Watcher.StartHeight != -1 {
		watcherDB.Store(config.Watcher.StartHeight, watcher.Block)
	}
	go watcherDB.ListenDB(syncCtx)
	log.Infoln("start processing functions")
	for i := 0; i < w.Worker; i++ {
		wg.Add(2)
		go processingBlock(syncCtx, wg, watcherDB.SubscribeBlock(), api)
		go processingTx(syncCtx, wg, watcherDB.SubscribeTx(), watcherDB.SubscribeTxHash(), api)
	}
	<-syncCtx.Done()
	log.Infoln("mbelt-cosmos-watcher gracefully stopped")
}

func processingBlock(ctx context.Context, wg *sync.WaitGroup, heightChan <-chan int64, api *client.ClientApi) {
	for {
		select {
		case height := <-heightChan:
			block := api.GetBlockRpc(ctx, height)
			if block == nil {
				continue
			}
			log.Infoln("new block -> ", block.Height)
			services.App().BlocksService().Push(block)
		case <-ctx.Done():
			wg.Done()
		}
	}
}

func processingTx(ctx context.Context, wg *sync.WaitGroup, heightChan <-chan int64, hashesChan <-chan string, api *client.ClientApi) {
	for {
		select {
		// case height := <-heightChan:
		// 	txs := api.GetTxsRpc(height)
		// 	for _, tx := range txs {
		// 		log.Infoln("new tx -> ", tx.Height)
		// 		services.App().TransactionsService().Push(tx)
		// 	}
		case hash := <-hashesChan:
			tx := api.GetTx(ctx, hash)
			if tx == nil {
				continue
			}
			log.Infoln("new tx hash-> ", tx.TxResponse.TxHash)
			services.App().TransactionsService().Push(tx)
		case <-ctx.Done():
			wg.Done()
		}
	}
}
