package messages

import (
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore"
	"github.com/tendermint/tendermint/types"
	log "log"
)

type Service struct {
	config *config.Config
	ds     *datastore.KafkaDatastore
}

func Init(config *config.Config, ds *datastore.KafkaDatastore) (*Service, error) {
	return &Service{
		config: config,
		ds:     ds,
	}, nil
}

func (s *Service) Push(tx *types.EventDataTx) {
	// Empty messages has panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("[MessagesService][Recover]", "Throw panic", r)
		}
	}()

	m := map[string]interface{}{}
	m[fmt.Sprintf("%X", tx.Tx.Hash())] = s.serialize(tx)

	s.ds.Push(datastore.TopicMessages, m)
}

func (s *Service) serialize(tx *types.EventDataTx) map[string]interface{} {
	result := map[string]interface{}{}
	return result
}
