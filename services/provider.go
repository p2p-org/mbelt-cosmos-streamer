package services

import (
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-cosmos-streamer/services/blocks"
	"github.com/p2p-org/mbelt-cosmos-streamer/services/transactions"
)

var (
	provider ServiceProvider
)

type ServiceProvider struct {
	blocksService       *blocks.Service
	transactionsService *transactions.Service
}

func (p *ServiceProvider) Init(config *config.Config, kafkaDs *datastore.KafkaDatastore, pgDs *pg.PgDatastore) error {
	var err error
	p.blocksService, err = blocks.Init(config, kafkaDs, pgDs)
	if err != nil {
		return err
	}

	p.transactionsService, err = transactions.Init(config, kafkaDs, pgDs)
	if err != nil {
		return err
	}

	return nil
}

func (p *ServiceProvider) BlocksService() *blocks.Service {
	return p.blocksService
}

func (p *ServiceProvider) TransactionsService() *transactions.Service {
	return p.transactionsService
}

func App() *ServiceProvider {
	return &provider
}
