package network

import (
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/gVar"
)
var readymask	[]byte
func shardProcess(){
	var beginShard	shard.Instance
	shard.StartFlag = true
	beginShard.GenerateSeed(&shard.PreviousSyncBlockHash)
	beginShard.Sharding(&shard.GlobalGroupMems, &shard.ShardToGlobal)
	shard.MyMenShard = shard.GlobalGroupMems[shard.MyGlobalID]
	//intilizeMaskBit(&readymask, (shard.NumMems+7)>>3,false)

	if shard.MyMenShard.Role == 1{
		MinerReadyProcess()
	}	else {
		LeaderReadyProcess(&shard.GlobalGroupMems)
	}

}
func LeaderReadyProcess(ms *[]shard.MemShard){
	//var readaddr string
	readyCount := 0

	for readyCount <= int(gVar.ShardSize/2) {
		<-readyCh
		readyCount++
		//setMaskBit((*ms)[GlobalAddrMapToInd[readaddr]].InShardId, cosi.Enabled, &readymask)
	}

}
func MinerReadyProcess(){
	myShard := shard.GlobalGroupMems[shard.MyGlobalID].Shard
	SendShardReadyMessage(shard.GlobalGroupMems[shard.ShardToGlobal[myShard][0]].Address, "shardReady", nil)
}


func SendShardReadyMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}


func HandleShardReady(addr string) {
	readyCh <- addr

}