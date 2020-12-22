package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/pg"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
)

type Consistency struct {
	Cmd *cobra.Command

	lastBlock int64
}

func (c *Consistency) Init(cfg *config.Config) {
	c.Cmd = &cobra.Command{
		Use:   "consistency",
		Short: "A consistency of cosmos's entities to PostgreSQL DB through Kafka",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			c.Start(cfg)
		},
	}
}

func (c *Consistency) Start(cfg *config.Config) {
	exitCode := 0
	defer os.Exit(exitCode)
	syncCtx, syncCancel := context.WithCancel(context.Background())

	go func() {
		var gracefulStop = make(chan os.Signal)
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		signal.Notify(gracefulStop, syscall.SIGHUP)

		sig := <-gracefulStop
		log.Infof("Caught sig: %+v", sig)
		log.Infoln("Wait for graceful shutdown to finish.")

		syncCancel()
	}()

	db, err := pg.Init(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	c.lastBlock = db.GetLastConsistencyBlock()

	go func() {
		for {
			select {
			case <-time.Tick(time.Second * 2):
				if c.lastBlock == 0 {
					c.lastBlock = db.GetMinBlockHeight()
				}
				blocks := db.GetBlocksWithCountTxs()
				var blockSetter int64
				for _, block := range blocks {
					log.Infoln(block.Height, c.lastBlock)
					if block.Height == c.lastBlock || block.Height == (c.lastBlock+1) {
						if block.CountTxs != block.NumTx {
							blockSetter = block.Height - 1
						} else {
							blockSetter = block.Height
							c.lastBlock++
						}
					} else {
						break
					}
				}
				db.SetConsistency(blockSetter)
				log.Infoln("set block consistency: ", blockSetter)
			}
		}
	}()
	<-syncCtx.Done()
	log.Infoln("mbelt-cosmos-consistency gracefully stopped")
}
