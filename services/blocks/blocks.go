package blocks

import (
	"fmt"

	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/utils"
	"github.com/prometheus/common/log"

	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/pg"
	"github.com/tendermint/tendermint/types"
)

type Service struct {
	config  *config.Config
	kafkaDs *datastore.KafkaDatastore
}

func Init(config *config.Config, kafkaDs *datastore.KafkaDatastore, pgDs *pg.PgDatastore) (*Service, error) {
	return &Service{
		config:  config,
		kafkaDs: kafkaDs,
	}, nil
}

func (s *Service) Push(block *types.Block) {

	defer func() {
		if r := recover(); r != nil {
			log.Infoln("[BlocksService][Recover]", "Throw panic", r)
		}
	}()
	m := map[string]interface{}{}

	m[block.Hash().String()] = s.serialize(block)
	s.kafkaDs.Push(datastore.TopicBlocks, m)
}

func (s *Service) serialize(block *types.Block) map[string]interface{} {
	txsHash := make([]string, 0)

	for _, tx := range block.Data.Txs {
		txsHash = append(txsHash, fmt.Sprintf("%X", tx.Hash()))
	}

	result := map[string]interface{}{
		"hash":            block.Hash().String(),
		"chain_id":        block.Header.ChainID,
		"height":          uint64(block.Header.Height),
		"time":            block.Header.Time.Unix(),
		"num_tx":          uint64(block.Header.NumTxs),
		"txs_hash":        utils.ToVarcharArray(txsHash),
		"total_txs":       uint64(block.Header.TotalTxs),
		"last_block_hash": block.Header.LastBlockID.Hash.String(),
		"validator":       block.Header.ValidatorsHash.String(),
		"status":          client.PendingStatus,
	}

	if uint64(block.Header.NumTxs) == 0 {
		result["status"] = client.ConfirmedStatus
	}

	return result
}
