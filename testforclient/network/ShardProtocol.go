package network

import (
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/gVar"
)

//syncMask indicates the sync of each shard
var syncmask []byte

//syncProcess processing sync after one epoch
func syncProcess(){
	//flag for all sync
	allflag := true
	timeoutflag := true
	//ask user i for sync, -1 means done, otherwise ask i+1
	aski := make([]int, gVar.ShardCnt)
	syncmask = make([]byte, (gVar.ShardCnt+7)>>3)
	for i := range syncmask {
		syncmask[i] = 0xff // all disabled
	}
	for i := 0; i < shard.NumMems; i++ {
		if !maskBit(i, &syncmask){
			go sendSyncMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][0]].Address, "requestsb", nil)
			aski[i] = 0
		}
	}
	for allflag {
		select  {
			case
		}
	}



}


// sendCosiMessage send cosi message
func sendSyncMessage(addr string, command string, message interface{}, ) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}