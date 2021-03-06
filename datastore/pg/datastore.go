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
									where t2.height is null order by t1.height limit 10000;`

const queryGetAllLostBlocks = `select generate_series as h 
			from generate_series((select min(height) from cosmos.blocks), (select max(height) - 1 from cosmos.blocks)) 
					left join cosmos.blocks on generate_series.generate_series = height where height IS NULL order by h limit 5000;`

const queryGetAllLostTransactions = `select distinct(b.height) from (select unnest(txs_hash) as tx_hash, height from cosmos.blocks where num_tx > 0 order by height) as b
    left join cosmos.transactions t on t.tx_hash = b.tx_hash where t.tx_hash is null order by b.height asc limit 2000`

const queryGetAllLostTransactionsHashes = `select b.tx_hash from (select unnest(txs_hash) as tx_hash, height from cosmos.blocks where num_tx > 0 order by height) as b
    left join cosmos.transactions t on t.tx_hash = b.tx_hash where t.tx_hash is null order by b.height asc limit 2000`

const queryBlocksWithCountTxs = `SELECT b.height, b.num_tx, count(t.block_height) as count_txs FROM cosmos.blocks b
                                                                        left join cosmos.transactions t on t.block_height = b.height and t.chain_id = b.chain_id
where height >= coalesce((select block_height from cosmos.consistency order by block_height desc limit 1), 0)
group by t.block_height, b.height,  num_tx order by b.height limit 50000`

type BlocksWithCountTxs struct {
	Height   int64
	NumTx    int64
	CountTxs int64
}

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
		defer rows.Close()
		for rows.Next() {
			var height int64
			if err := rows.Scan(&height); err != nil {
				log.Errorln(err)
				continue
			}
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
		defer rows.Close()
		for rows.Next() {
			var height int64
			if err := rows.Scan(&height); err != nil {
				log.Errorln(err)
				continue
			}
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
		defer rows.Close()
		for rows.Next() {
			var height int64
			if err := rows.Scan(&height); err != nil {
				log.Errorln(err)
				continue
			}
			heights = append(heights, height)
		}
	}

	return heights
}

func (ds *PgDatastore) GetAllLostTransactionsHashes() []string {
	var hashes []string
	rows, err := ds.conn.Query(queryGetAllLostTransactionsHashes)
	if err != nil {
		log.Errorln(err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var hash string
			if err := rows.Scan(&hash); err != nil {
				log.Errorln(err)
				continue
			}
			hashes = append(hashes, hash)
		}
	}

	return hashes
}

func (ds *PgDatastore) GetLastConsistencyBlock() (height int64) {
	row := ds.conn.QueryRow("select coalesce((select block_height from cosmos.consistency order by block_height desc limit 1), (select min(height) from cosmos.blocks))")
	row.Scan(&height)
	return height
}

func (ds *PgDatastore) GetMinBlockHeight() (height int64) {
	row := ds.conn.QueryRow("select min(height) from cosmos.blocks")
	if err := row.Scan(&height); err != nil {
		log.Errorln(err)
	}
	return height
}

func (ds *PgDatastore) GetBlocksWithCountTxs() []BlocksWithCountTxs {
	var stats []BlocksWithCountTxs
	rows, _ := ds.conn.Query(queryBlocksWithCountTxs)
	for rows.Next() {
		var item BlocksWithCountTxs
		rows.Scan(&item.Height, &item.NumTx, &item.CountTxs)
		stats = append(stats, item)
	}
	rows.Close()
	return stats
}

func (ds *PgDatastore) SetConsistency(height int64) {
	ds.conn.Exec("select cosmos.set_consistency($1::bigint)", height)
}

func (ds *PgDatastore) BlockStatusChange(height int64) {
	ds.conn.Exec("select cosmos.block_status_change($1)", height)
}
