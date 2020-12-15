package commands

import (
	"log"

	"github.com/jinzhu/configor"
	"github.com/p2p-org/mbelt-cosmos-streamer/cmd/worker"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
)

var (
	cfg            config.Config
	watcherWorker  = worker.Watcher{}
	streamerWorker = worker.Streamer{}
)

func init() {
	if err := configor.Load(&cfg); err != nil {
		log.Fatal(err)
	}
	streamerWorker.Init(&cfg)
	watcherWorker.Init(&cfg)

	streamerWorker.Cmd.AddCommand(watcherWorker.Cmd)
}

func Execute() error {
	return streamerWorker.Cmd.Execute()
}
