package network

import (
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/Reputation"
)
//var currentTxList *[]basic.TxList
//var currentTxDecSet *[]basic.TxDecSet

func RepProcess(ms *[]shard.MemShard) {
	go MyRepBlockChain.MineRepBlock(ms)
	switch
}
