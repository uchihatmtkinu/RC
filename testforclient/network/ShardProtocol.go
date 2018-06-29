package network

import (
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/gVar"
	"fmt"
)
var readymask	[]byte
func ShardProcess(){
	var beginShard	shard.Instance
	CurrentEpoch++
	shard.StartFlag = true
	readyCh = make(chan string)
	shard.ShardToGlobal = make([][]int, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		shard.ShardToGlobal[i] = make([]int, gVar.ShardSize)
	}
	beginShard.GenerateSeed(&shard.PreviousSyncBlockHash)
	beginShard.Sharding(&shard.GlobalGroupMems, &shard.ShardToGlobal)
	//shard.MyMenShard = &shard.GlobalGroupMems[MyGlobalID]

	LeaderAddr = shard.GlobalGroupMems[shard.ShardToGlobal[shard.MyMenShard.Shard][0]].Address


	if shard.MyMenShard.Role == 1{
		MinerReadyProcess()
	}	else {
		LeaderReadyProcess(&shard.GlobalGroupMems)
	}
	fmt.Println("shard finished")
}
func LeaderReadyProcess(ms *[]shard.MemShard){
	//var readaddr string
	readyCount := 1
	//fmt.Println("wait for ready")
	//TODO modify the number of shardsize
	for readyCount < int(gVar.ShardSize/2) {
		<-readyCh
		readyCount++
		//fmt.Println("ReadyGet")
		//setMaskBit((*ms)[GlobalAddrMapToInd[readaddr]].InShardId, cosi.Enabled, &readymask)
	}


}
func MinerReadyProcess(){

	SendShardReadyMessage(LeaderAddr, "shardReady", "shardReady")
}


func SendShardReadyMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}


func HandleShardReady(addr string) {
	readyCh <- addr

}