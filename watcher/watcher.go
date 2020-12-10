package watcher

import (
	"time"

	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/p2p-org/mbelt-cosmos-streamer/datastore/pg"
)

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
	w.GetAllLostBlock()
	return nil
}

func (w *Watcher) ListenDB() {
	for {
		heights := w.db.GetLostBlocks()
		for _, height := range heights {
			w.Store(height)
		}
		time.Sleep(time.Minute)
	}
}

func (w *Watcher) GetAllLostBlock() {
	heights := w.db.GetAllLostBlocks()
	for _, height := range heights {
		w.Store(height)
	}
}
