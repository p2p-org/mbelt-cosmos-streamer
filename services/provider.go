package services

import (
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-cosmos-streamer/services/blocks"
)

var (
	provider ServiceProvider
)

type ServiceProvider struct {
	blocksService *blocks.Service
	// messagesService  *messages.MessagesService
	// processorService *processor.ProcessorService
}

func (p *ServiceProvider) Init(config *config.Config, kafkaDs *datastore.KafkaDatastore, pgDs *pg.PgDatastore) error {
	var err error
	p.blocksService, err = blocks.Init(config, kafkaDs, pgDs)
	if err != nil {
		return err
	}

	// p.messagesService, err = messages.Init(config, kafkaDs, apiClient)
	//
	// if err != nil {
	// 	return err
	// }
	//
	// p.processorService, err = processor.Init(config, kafkaDs, apiClient)
	//
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (p *ServiceProvider) BlocksService() *blocks.Service {
	return p.blocksService
}

//
// func (p *ServiceProvider) MessagesService() *messages.MessagesService {
// 	return p.messagesService
// }
//
// func (p *ServiceProvider) ProcessorService() *processor.ProcessorService {
// 	return p.processorService
// }

func App() *ServiceProvider {
	return &provider
}
