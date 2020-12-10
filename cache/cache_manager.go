package cache

import (
	"context"
	"fmt"

	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/prometheus/common/log"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

type CacheManager interface {
	Init(config *config.Config) error
	Close() error
	StoreBlock(ctx context.Context, height int64, block *types.Block)
	GetBlock(ctx context.Context, height int64) *types.Block
}

type CacheApi struct {
	client.Api
	cache CacheManager
}

func (c *CacheApi) Init(cfg *config.Config) error {
	c.cache = &CacheRedis{}
	if err := c.cache.Init(cfg); err != nil {
		return err
	}

	return c.Api.Init(cfg)
}

func (c *CacheApi) SubscribeBlock(ctx context.Context) chan ctypes.ResultEvent {
	blocks := make(chan ctypes.ResultEvent)
	go func() {
		for block := range c.Api.SubscribeBlock(ctx) {
			newBlock := block.Data.(types.EventDataNewBlock)
			log.Infoln("cache blocks -> ", block.Data.(types.EventDataNewBlock).Block.Height)
			c.cache.StoreBlock(ctx, newBlock.Block.Height, newBlock.Block)
			blocks <- block
		}
	}()
	return blocks
}

func (c *CacheApi) SubscribeTxs(ctx context.Context) chan ctypes.ResultEvent {
	txs := make(chan ctypes.ResultEvent)
	go func() {
		for tx := range c.Api.SubscribeTxs(ctx) {
			newTx := tx.Data.(types.EventDataTx)
			log.Infoln("cache tx -> ", fmt.Sprintf("%X", newTx.Tx.Hash()))
			txs <- tx
		}
	}()
	return txs
}

func (c *CacheApi) GetBlock(ctx context.Context, height int64) *types.Block {
	if block := c.cache.GetBlock(ctx, height); block != nil {
		return block
	}
	log.Infoln("get block from http ->", height)

	block := c.Api.GetBlock(height)
	c.cache.StoreBlock(ctx, height, block)

	return block
}

func (c *CacheApi) Stop() {
	if err := c.cache.Close(); err != nil {
		log.Errorln(err)
	}
	c.Api.Stop()
}
