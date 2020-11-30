package cache

import (
	"context"
	"fmt"
	"sync"

	"github.com/gomodule/redigo/redis"
	"github.com/p2p-org/mbelt-cosmos-streamer/client"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/prometheus/common/log"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

var cache CacheManager

type CacheApi struct {
	client.Api
}

func (c *CacheApi) Init(cfg *config.Config) error {
	cache = CacheManager{}
	return c.Api.Init(cfg)
}

func (c *CacheApi) SubscribeBlock(ctx context.Context) chan ctypes.ResultEvent {
	blocks := make(chan ctypes.ResultEvent)
	go func() {
		for block := range c.Api.SubscribeBlock(ctx) {
			newBlock := block.Data.(types.EventDataNewBlock)
			log.Infoln("cache blocks -> ", block.Data.(types.EventDataNewBlock).Block.Height)
			cache.StoreBlock(newBlock.Block.Height, newBlock)
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

func (c *CacheApi) GetBlock(height int64) *types.Block {
	if block, ok := cache.blocksSync.Load(height); ok {
		log.Infoln("get block from cache -> ", height)
		return block.(types.EventDataNewBlock).Block
	}
	log.Infoln("get block from http ->", height)
	// TODO save block to cache
	return c.Api.GetBlock(height)
}

type CacheManager struct {
	pool       *redis.Pool
	conn       redis.Conn
	blocksSync sync.Map
}

//
// func InitCacheManager() (*CacheManager, error) {
// 	manager := &CacheManager{
// 		// pool: &redis.Pool{
// 		// 	// Maximum number of idle connections in the pool.
// 		// 	MaxIdle: 80,
// 		// 	// max number of connections
// 		// 	MaxActive: 12000,
// 		// 	// Dial is an application supplied function for creating and
// 		// 	// configuring a connection.
// 		// 	Dial: func() (redis.Conn, error) {
// 		// 		c, err := redis.Dial("tcp", ":6379")
// 		// 		if err != nil {
// 		// 			panic(err.Error())
// 		// 		}
// 		// 		return c, err
// 		// 	},
// 		// },
// 	}
//
// 	manager.CachBlock = make(chan node.Block, 10000)
// 	// manager.conn = manager.pool.Get()
// 	// redis.Strings()
// 	return manager, nil
// }
//
// func (cm *CacheManager) ProcesingBlocks(blockChan <-chan node.Block) {
// 	for i := range blockChan {
// 		if i.NumTxs == 0 && i.Status == node.PendingStatus {
// 			i.Status = node.ConfirmedStatus
// 			// log.Infoln(i)
// 			cm.CachBlock <- i
// 		} else {
// 			cachBlock := cachingBlock{
// 				procesedTx: i.NumTxs,
// 				block:      i,
// 			}
// 			cm.blocksSync.Store(int64(i.Height), cachBlock)
// 			// cm.blocks[int64(i.Height)] = cachBlock
// 			// log.Infoln(i)
// 			cm.CachBlock <- i
// 		}
// 	}
// }
func (cm *CacheManager) StoreBlock(height int64, block types.EventDataNewBlock) {
	cm.blocksSync.Store(height, block)
}

// func (cm *CacheManager) ProcesingTxs(txsChan chan node.TransactionWithBlockInfo) {
// 	for i := range txsChan {
// 		block, ok := cm.blocksSync.Load(i.BlockNum)
// 		if ok && block.(cachingBlock).block.Txs[i.TxNum].TxHash == i.Tx.TxHash {
// 			block := block.(cachingBlock)
// 			block.block.Txs[i.TxNum] = i.Tx
// 			if block.procesedTx == 1 {
// 				block.block.Status = node.ConfirmedStatus
// 				cm.CachBlock <- block.block
// 				cm.blocksSync.Delete(i.BlockNum)
// 			}
// 			log.Infof("\n\nblocks:\n %v\n\n", block)
// 			block.procesedTx -= 1
// 			cm.blocksSync.Store(i.BlockNum, block)
//
// 		} else {
// 			// log.Infoln("Errrrrororororororororo \n\n\n\n", i, " \n\n\nerer\n\n")
// 			go resendTxToChan(txsChan, i)
// 		}
//
// 	}
// }
//
// func resendTxToChan(txsChan chan node.TransactionWithBlockInfo, Tx node.TransactionWithBlockInfo) {
// 	time.Sleep(time.Second)
// 	txsChan <- Tx
// }
//
// func (cm *CacheManager) Close() {
// 	cm.conn.Close()
// }
