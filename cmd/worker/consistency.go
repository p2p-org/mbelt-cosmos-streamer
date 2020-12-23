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
				blocks := db.GetBlocksWithCountTxs()
				var blockSetter int64
				for _, block := range blocks {
					if block.Height == c.lastBlock {
						if block.CountTxs != block.NumTx {
							blockSetter = block.Height - 1
						} else {
							blockSetter = block.Height
							db.BlockStatusChange(block.Height)
							c.lastBlock++
						}
					} else {
						break
					}
				}
				c.lastBlock = blockSetter
				db.SetConsistency(blockSetter)
				log.Infoln("set block consistency: ", blockSetter)
			}
		}
	}()
	<-syncCtx.Done()
	log.Infoln("mbelt-cosmos-consistency gracefully stopped")
}
