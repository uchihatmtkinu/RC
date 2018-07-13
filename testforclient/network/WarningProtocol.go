package network

import (
	"github.com/uchihatmtkinu/RC/shard"
	"time"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
)

type WarningInfo struct {
	Epoch	int
	Height	int
	ID 		int
}
var WarningCh chan WarningInfo
var legalmask []byte
var currentLeader int
func WarningProcess(){
	var it *shard.MemShard
	//TODO place it in the shardProtocol
	intilizeMaskBit(&legalmask, (int(gVar.ShardSize)+7)>>3, cosi.Disabled)

	numWarning := 1
	for numWarning <= shard.NumMems {
		select {
		case x := <-WarningCh:
			{
				if x.Height == Height {
					numWarning++
					setMaskBit(shard.GlobalGroupMems[MyGlobalID].InShardId, cosi.Enabled, &legalmask)
				}
			}
		case <-time.After(timeoutSync): {
			for i := uint32(1); i < gVar.ShardSize; i++ {
				it = &shard.GlobalGroupMems[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
				if it.Legal == 0 && maskBit(it.InShardId, &legalmask) == cosi.Disabled {
					SendCosiMessage(it.Address, "warning", WarningInfo{CurrentEpoch, Height, MyGlobalID})
				}
			}
		}
		}
	}
	shard.GlobalGroupMems[currentLeader].Legal = 1
	shard.GlobalGroupMems[currentLeader].Role = 1
	currentLeader++
	shard.GlobalGroupMems[currentLeader].Role = 0
	LeaderAddr = shard.GlobalGroupMems[currentLeader].Address
}
