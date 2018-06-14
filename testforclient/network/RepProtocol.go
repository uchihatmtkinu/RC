package network

import (
	"github.com/uchihatmtkinu/RC/shard"
)
//var currentTxList *[]basic.TxList
//var currentTxDecSet *[]basic.TxDecSet

func RepProcess(ms *[]shard.MemShard) {
	go MyRepBlockChain.MineRepBlock(ms)
	switch
}
