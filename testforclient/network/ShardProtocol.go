package network

import (
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/gVar"
)

func shardProcess(){
	var beginShard	shard.Instance
	shard.StartFlag = true
	beginShard.GenerateSeed(&shard.PreviousSyncBlockHash)
	beginShard.Sharding(&shard.GlobalGroupMems, &shard.ShardToGlobal)
	shard.MyMenShard = shard.GlobalGroupMems[shard.MyGlobalID]
	if shard.MyMenShard.Role == 1{
		minerReadyProcess()
	}	else {
		leaderReadyProcess()
	}

}
func leaderReadyProcess(){
	readyCount := 0
	for readyCount <= int(gVar.ShardSize/2) {
		<-readyCh
		readyCount++
	}

}
func minerReadyProcess(){
	myShard := shard.GlobalGroupMems[shard.MyGlobalID].Shard
	sendShardReadyMessage(shard.GlobalGroupMems[shard.ShardToGlobal[myShard][0]].Address, "shardReady", nil)
}


func sendShardReadyMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}


