package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/prometheus/common/log"
	"github.com/tendermint/tendermint/types"
)

type CacheRedis struct {
	CacheManager
	client         *redis.Client
	cfg            *config.Config
	expirationTime time.Duration
}

func (cr *CacheRedis) Init(cfg *config.Config) (err error) {
	cr.cfg = cfg
	cr.client = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	cr.expirationTime, err = time.ParseDuration(cfg.Redis.ExpirationCache)
	return err
}

func (cr *CacheRedis) StoreBlock(ctx context.Context, height int64, block *types.Block) {
	log.Infoln("redis cache block")
	data, err := json.Marshal(block)
	if err != nil {
		log.Errorln(err)
	}
	cr.client.Set(ctx, cr.getBlockKey(height), data, cr.expirationTime)
}

func (cr *CacheRedis) GetBlock(ctx context.Context, height int64) *types.Block {
	str, err := cr.client.Get(ctx, cr.getBlockKey(height)).Result()
	if err != nil {
		log.Errorln(err)
		return nil
	}
	var block *types.Block
	if err := json.Unmarshal([]byte(str), &block); err != nil {
		log.Errorln(err)
	}
	return block
}

func (cr *CacheRedis) Close() error {
	return cr.client.Close()
}

func (cr *CacheRedis) getBlockKey(height int64) string {
	return fmt.Sprintf("blocks:%d", height)
}
