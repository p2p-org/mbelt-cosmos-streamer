package services

import (
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/pg"
)

func InitServices(config *config.Config) error {
	kafkaDs, err := datastore.Init(config)
	if err != nil {
		return err
	}
	pgDs, err := pg.Init(config)
	if err != nil {
		return err
	}

	if err = provider.Init(config, kafkaDs, pgDs); err != nil {
		return err
	}

	return nil
}
