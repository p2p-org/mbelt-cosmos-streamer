package watcher

type CacheWatcher struct {
	heights           []int64
	chanHeightsBlocks chan int64
	chanHeightsTxs    chan int64
}

func (c *CacheWatcher) InitCache() {
	c.chanHeightsBlocks = make(chan int64, 1000000)
	c.chanHeightsTxs = make(chan int64, 1000000)
}

func (c *CacheWatcher) Store(height int64, object string) {
	switch object {
	case Tx:
		c.chanHeightsTxs <- height
	case Block:
		c.chanHeightsBlocks <- height
	}
}

func (c *CacheWatcher) SubscribeBlock() <-chan int64 {
	return c.chanHeightsBlocks
}

func (c *CacheWatcher) SubscribeTx() <-chan int64 {
	return c.chanHeightsTxs
}
