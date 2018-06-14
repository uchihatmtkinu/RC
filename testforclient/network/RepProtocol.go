package network

import (
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/shard"
)
var currentTxList *[]basic.TxList
var currentTxDecSet *[]basic.TxDecSet

func RepProcess(ms []shard.MemShard) {
	for _, item :=  range *currentTxList {
		for j,itTx:=range item.TxArray{
			if (itTx.Hash)
		}
	}
}
