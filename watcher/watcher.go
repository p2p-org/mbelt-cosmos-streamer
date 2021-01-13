package watcher

import (
	"context"
	"time"

	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/pg"
	"github.com/prometheus/common/log"
)

const Tx = "tx"
const Block = "block"

type Watcher struct {
	CacheWatcher
	db *pg.PgDatastore
}

func (w *Watcher) Init(cfg *config.Config) error {
	var err error
	if w.db, err = pg.Init(cfg); err != nil {
		return err
	}
	w.InitCache()
	w.GetAllLostBlocks()
	w.GetAllLostTxs()
	return nil
}

func (w *Watcher) ListenDB(ctx context.Context) {
	timer := time.NewTicker(time.Second * 15)

	for {
		select {
		case <-timer.C:
			log.Infoln("get GetLostBlocks")
			heights := w.db.GetLostBlocks()
			for _, height := range heights {
				w.Store(height, Block)
				w.Store(height, Tx)
			}
			log.Infoln("get GetAllLostTransactions")

			heights = w.db.GetAllLostTransactions()
			for _, height := range heights {
				w.Store(height, Block)
				w.Store(height, Tx)
			}
			log.Infoln("get GetAllLostBlocks")

			heights = w.db.GetAllLostBlocks()
			for _, height := range heights {
				w.Store(height, Block)
				w.Store(height, Tx)
			}
		case <-ctx.Done():
			timer.Stop()
			break
		}
	}
}

func (w *Watcher) GetAllLostBlocks() {
	heights := w.db.GetAllLostBlocks()
	for _, height := range heights {
		w.Store(height, Block)
	}
}

func (w *Watcher) GetAllLostTxs() {
	heights := w.db.GetAllLostBlocks()
	for _, height := range heights {
		w.Store(height, Tx)
	}
}
