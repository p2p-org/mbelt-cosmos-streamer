package commands

import (
	"github.com/jinzhu/configor"
	"github.com/p2p-org/mbelt-cosmos-streamer/cmd/worker"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/prometheus/common/log"
)

var (
	cfg               config.Config
	watcherWorker     = worker.Watcher{}
	streamerWorker    = worker.Streamer{}
	consistencyWorker = worker.Consistency{}
)

func init() {
	if err := configor.Load(&cfg); err != nil {
		log.Fatal(err)
	}
	if err := log.Base().SetLevel(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}

	streamerWorker.Init(&cfg)
	watcherWorker.Init(&cfg)
	consistencyWorker.Init(&cfg)

	streamerWorker.Cmd.AddCommand(watcherWorker.Cmd)
	streamerWorker.Cmd.AddCommand(consistencyWorker.Cmd)
}

func Execute() error {
	return streamerWorker.Cmd.Execute()
}
