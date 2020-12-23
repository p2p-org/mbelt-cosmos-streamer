package main

import (
	"database/sql"
	"time"

	"github.com/p2p-org/mbelt-cosmos-streamer/watcher"
	"github.com/prometheus/common/log"

	"github.com/lib/pq"
)

const sync_delay = 10 * time.Second
const idle_delay = 5 * time.Second
const scrn_width = 65
const min_reconn = 5 * time.Second
const max_reconn = time.Minute
const pgString = "postgres://sink:Ekj31R2_03S2IwLoPsWVa28_sMx_xoS@192.168.1.161:5432/raw?sslmode=disable"

const channel = "events"

var db *sql.DB

var cache watcher.CacheWatcher

func main() {
	cache.Init()
	// var err error
	// db, err = sql.Open("postgres", pgString)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer db.Close()
	// go notifyNewBlock()
	// go getLostBlocks()

	cache.Store(123123)
	time.Sleep(time.Second * 15)
	cache.Store(123124)

	for {
		select {
		case <-time.Tick(time.Second * 5):
			log.Infoln("tick ---- ")
			log.Infoln(cache.Count())
			if cache.Count() == 0 {
				log.Infoln(<-cache.Subscribe())
				log.Infoln(<-cache.Subscribe())
			}
		}
	}

	// s := make(chan struct{})
	// <-s
}

func notifyNewBlock() {
	reportErr := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Infof("Failed to start listener: %s", err)
		}
	}

	listener := pq.NewListener(pgString, min_reconn, max_reconn, reportErr)
	err := listener.Listen(channel)
	if err != nil {
		log.Infof(
			"Failed to LISTEN to channel '%s': %s",
			channel, err)
		panic(err)
	}
	log.Infof("Listening to notifications on channel \"%s\"", channel)

	for {
		select {
		case n := <-listener.Notify:
			log.Infoln("new notification -> ", n.Extra)
			// var c Counter
			// err := json.Unmarshal([]byte(n.Extra), &c)
			//
			// if err != nil {
			// 	log.Printf("Failed to parse '%s': %s", n.Extra, err)
			// } else {
			// 	log.Printf("Received event: %s", n.Extra)
			// 	updateCounter(cache, c)

		case <-time.After(idle_delay):
			if err := listener.Ping(); err != nil {
				log.Errorln(err)
			}
			return
		}
	}
}

func getLostBlocks() {
	log.Infoln(db.Ping())

	query := `SELECT t1.height + 1 as d
				FROM cosmos.blocks AS t1
         			LEFT JOIN cosmos.blocks AS t2 ON t2.height = t1.height + 1
				where t2.height is null;`

	for {
		rows, err := db.Query(query)
		if err != nil {
			log.Errorln(err)

		} else {
			var heights []int64
			rows.Scan(&heights)

			// for rows.Next() {
			// 	var height int64
			// 	rows.Scan(&height)
			// 	log.Infoln(height)
			// }
		}

		time.Sleep(time.Minute)
	}
}

func t() {
	d := make(map[int64]int64)
	d[12323] = 123123

}
