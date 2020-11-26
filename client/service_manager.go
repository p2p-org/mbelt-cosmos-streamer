package client

import (
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/tendermint/tendermint/types"
)

type Manager struct {
	nodeConnector *NodeManager
	// eventManager  *events.EventManager
	// cacheManager  *cache.CacheManager
	BlockProcessing func(*types.EventDataNewBlock)
	TxProcessing    func(tx *types.EventDataTx)
}

func (m *Manager) Start(cfg *config.Config) error {
	var err error

	m.nodeConnector, err = InitNodeManager(cfg)
	if err != nil {
		return err
	}

	err = m.nodeConnector.Connect()
	if err != nil {
		return err
	}

	if m.BlockProcessing != nil {
		go m.nodeConnector.SubscribeBlock(m.BlockProcessing)
	}
	if m.TxProcessing != nil {
		go m.nodeConnector.SubscribeTxs(m.TxProcessing)
	}
	// go m.cacheManager.ProcesingBlocks(m.nodeConnector.BlockChan)
	// go m.cacheManager.ProcesingTxs(m.nodeConnector.TxChan)
	// m.nodeConnector.ResendBlock(3108949)
	// go m.nodeConnector.CheckStatus()

	// go m.everyDayCheckLostBlock()

	// go m.reloadLostBlock()
	return nil
}

//
// func (m *Manager) reloadLostBlock() {
// 	for i := range m.eventManager.LostBlockChannel {
// 		log.Infoln(i)
// 		m.nodeConnector.ResendBlock(i)
// 	}
// }
//
// func (m *Manager) PrintIncomingBlock() {
// 	for i := range m.cacheManager.CachBlock {
// 		log.Errorln("lost block", i)
// 		m.eventManager.EmitNewBlock(i)
// 	}
//
// }
//
// func (m *Manager) everyDayCheckLostBlock() {
// 	for {
// 		err := m.eventManager.EmitCheckLostBlocks()
// 		if err != nil {
// 			log.Errorf("Error on emit find lost block on database, err: %v", err)
// 		}
// 		time.Sleep(time.Minute * 15)
// 	}
// }

func (m *Manager) Stop() {
	m.nodeConnector.Stop()
	// m.eventManager.Stop()
}
