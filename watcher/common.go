package watcher

type CacheWatcher struct {
	heights           []int64
	chanHeightsBlocks chan int64
	chanHeightsTxs    chan int64
	chanHashTxs       chan string
}

func (c *CacheWatcher) InitCache() {
	c.chanHeightsBlocks = make(chan int64, 1000000)
	c.chanHeightsTxs = make(chan int64, 1000000)
}

func (c *CacheWatcher) Store(value interface{}, object string) {
	switch object {
	case Tx:
		c.chanHeightsTxs <- value.(int64)
	case Block:
		c.chanHeightsBlocks <- value.(int64)
	case TxHash:
		c.chanHashTxs <- value.(string)
	}
}

func (c *CacheWatcher) SubscribeBlock() <-chan int64 {
	return c.chanHeightsBlocks
}

func (c *CacheWatcher) SubscribeTx() <-chan int64 {
	return c.chanHeightsTxs
}

func (c *CacheWatcher) SubscribeTxHash() <-chan string {
	return c.chanHashTxs
}
