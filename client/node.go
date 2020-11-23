package client

import (
	"time"

	"github.com/tendermint/tendermint/rpc/client"
)

const (
	blockQuery string = "tm.event = 'NewBlock'"
	txQuery    string = "tm.event = 'Tx'"
)

type StatusEnum string

const (
	PendingStatus   StatusEnum = "pending"
	ConfirmedStatus StatusEnum = "confirmed"
	RejectedStatus  StatusEnum = "rejected"
	OnForkStatus    StatusEnum = "onfork"
)

type BlockConnector struct {
	lcdURL    string
	wsClient  *client.HTTP
	BlockChan chan<- Block
}

type TransactionConnector struct {
	lcdUrl   string
	wsClient *client.HTTP
	TxChan   chan<- Transaction
}

// Block info
type Block struct {
	Hash          string             `json:"hash"`
	ChainID       string             `json:"chainID"`
	Height        uint64             `json:"height"`
	Time          time.Time          `json:"time"`
	NumTxs        uint64             `json:"numTx"`
	TotalTxs      uint64             `json:"totalTxs"`
	LastBlockHash string             `json:"lastBlockHash"`
	Validators    string             `json:"validators"`
	Status        StatusEnum         `json:"status"`
	Txs           []BlockTransaction `json:"txs"`
}

type TransactionWithBlockInfo struct {
	BlockNum int64
	TxNum    uint32
	Tx       BlockTransaction
}

type BlockTransaction struct {
	TxHash     string
	Messages   []Message
	Fee        string
	Signatures string
	Memo       string
	Status     StatusEnum
}

type Message struct {
	MsgType      string `json:"type"`
	Msg          string `json:"value"`
	Events       string `json:"events"`
	ExternalInfo string `json:"external_info"`
	Log          string `json:"log"`
}

type Transaction struct {
	Height uint64
	Index  uint32
	Tx     []byte // это сырая транзакция, ее можем не слать, это по желанию
	Hash   string
	Result struct {
		Log       string
		GasWanted int64
		GasUsed   int64
		Events    []struct {
			Type       string
			Attributes []struct {
				Key   string
				Value string
			}
		}
		Codespace string
	}
}

type TempMessage struct {
	MsgType string                 `json:"type"`
	Msg     map[string]interface{} `json:"value"`
}

type Fee struct {
	GasWanted string `json:"gas_wanted"`
	GasUsed   string `json:"gas_used"`
	Amount    string `json:"gas_amount"`
}
