package pg

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/prometheus/common/log"
)

const QueryGetMaxHeight = `SELECT coalesce(max(height), 0) FROM cosmos.blocks`

const queryGetLostBlocks = `SELECT t1.height + 1 as d FROM cosmos.blocks AS t1
                                   LEFT JOIN cosmos.blocks AS t2 ON t2.height = t1.height + 1
									where t2.height is null order by t1.height asc offset 1  limit 10000;`

const queryGetAllLostBlocks = `select generate_series as h 
			from generate_series((select min(height) from cosmos.blocks), (select max(height) - 1 from cosmos.blocks)) 
					left join cosmos.blocks on generate_series.generate_series = height where height IS NULL order by h limit 10000;`

const queryGetAllLostTransactions = `select b.height from (select unnest(txs_hash) as tx_hash, height from cosmos.blocks where num_tx > 0 limit 100000) as b
    left join cosmos.transactions t on t.tx_hash = b.tx_hash where t.tx_hash is null order by b.height`

type PgDatastore struct {
	conn *sql.DB
}

func Init(config *config.Config) (*PgDatastore, error) {
	if config == nil {
		return nil, errors.New("can't init postgres datastore with nil config")
	}
	db, err := sql.Open("postgres", config.PgUrl)

	ds := &PgDatastore{db}
	return ds, err
}

func (ds *PgDatastore) GetMaxHeight() (height int, err error) {
	r := ds.conn.QueryRow(QueryGetMaxHeight)
	err = r.Scan(&height)
	return
}

func (ds *PgDatastore) GetLostBlocks() []int64 {
	var heights []int64
	rows, err := ds.conn.Query(queryGetLostBlocks)
	if err != nil {
		log.Errorln(err)

	} else {
		for rows.Next() {
			var height int64
			rows.Scan(&height)
			heights = append(heights, height)
		}
	}

	return heights
}

func (ds *PgDatastore) GetAllLostBlocks() []int64 {
	var heights []int64
	rows, err := ds.conn.Query(queryGetAllLostBlocks)
	if err != nil {
		log.Errorln(err)

	} else {
		for rows.Next() {
			var height int64
			rows.Scan(&height)
			heights = append(heights, height)
		}
	}

	return heights
}

func (ds *PgDatastore) GetAllLostTransactions() []int64 {
	var heights []int64
	rows, err := ds.conn.Query(queryGetAllLostTransactions)
	if err != nil {
		log.Errorln(err)

	} else {
		for rows.Next() {
			var height int64
			rows.Scan(&height)
			heights = append(heights, height)
		}
	}

	return heights
}
