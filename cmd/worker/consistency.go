package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
)

type Consistency struct {
	Cmd *cobra.Command

	lastBlock int64
}

func (c *Consistency) Init() {
	c.Cmd = &cobra.Command{
		Use:   "consistency",
		Short: "A consistency of cosmos's entities to PostgreSQL DB through Kafka",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			c.Start()
		},
	}
}

func (c *Consistency) Start() {
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

	<-syncCtx.Done()
	log.Infoln("mbelt-cosmos-consistency gracefully stopped")
}
