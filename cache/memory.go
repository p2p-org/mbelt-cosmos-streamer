package cache

import (
	"context"
	"sync"

	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/prometheus/common/log"
	"github.com/tendermint/tendermint/types"
)

type CacheMemory struct {
	CacheManager
	blocksSync sync.Map
}

func (cm *CacheMemory) Init(cfg *config.Config) error {
	return nil
}

func (cm *CacheMemory) GetBlock(ctx context.Context, height int64) *types.Block {
	if block, ok := cm.blocksSync.Load(height); ok {
		log.Infoln("get block from cache -> ", height)
		return block.(*types.Block)
	}
	return nil
}

func (cm *CacheMemory) StoreBlock(ctx context.Context, height int64, block *types.Block) {
	cm.blocksSync.Store(height, block)
}
