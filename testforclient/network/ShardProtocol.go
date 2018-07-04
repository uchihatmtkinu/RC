package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

var readymask []byte

func ShardProcess() {
	var beginShard shard.Instance

	Reputation.CurrentRepBlock.Mu.Lock()
	Reputation.CurrentRepBlock.Round = -1
	Reputation.CurrentRepBlock.Mu.Unlock()

	shard.StartFlag = true
	shard.ShardToGlobal = make([][]int, gVar.ShardCnt)
	if CurrentEpoch != 1 {
		for i := 0; i < int(gVar.ShardSize*gVar.ShardCnt); i++ {
			shard.GlobalGroupMems[i].ClearRep()
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		shard.ShardToGlobal[i] = make([]int, gVar.ShardSize)

	}
	beginShard.GenerateSeed(&shard.PreviousSyncBlockHash)
	beginShard.Sharding(&shard.GlobalGroupMems, &shard.ShardToGlobal)
	//shard.MyMenShard = &shard.GlobalGroupMems[MyGlobalID]

	LeaderAddr = shard.GlobalGroupMems[shard.ShardToGlobal[shard.MyMenShard.Shard][0]].Address
	CacheDbRef.ShardNum = uint32(shard.MyMenShard.Shard)
	CacheDbRef.Leader = uint32(shard.ShardToGlobal[shard.MyMenShard.Shard][0])

	if shard.MyMenShard.Role == 1 {
		MinerReadyProcess()
	} else {
		LeaderReadyProcess(&shard.GlobalGroupMems)
	}
	fmt.Println("shard finished")
}

//LeaderReadyProcess leader use this
func LeaderReadyProcess(ms *[]shard.MemShard) {
	var readyMessage readyInfo
	var it *shard.MemShard
	var responsemask []byte

	intilizeMaskBit(&responsemask, (int(gVar.ShardSize)+7)>>3, cosi.Disabled)

	readyCount := 1
	//sent announcement
	for i := 1; i < int(gVar.ShardSize); i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		SendCosiMessage(it.Address, "readyAnnoun", readyInfo{MyGlobalID, CurrentEpoch})
	}
	fmt.Println("sent CoSi announce")
	//fmt.Println("wait for ready")
	//TODO modify int(gVar.ShardSize)/2
	for readyCount < int(gVar.ShardSize) {
		select {
		case readyMessage = <-readyCh:
			if readyMessage.Epoch == CurrentEpoch {
				readyCount++
				setMaskBit((*ms)[readyMessage.ID].InShardId, cosi.Enabled, &responsemask)
				fmt.Println("ReadyCount: ", readyCount)
			}
		case <-time.After(timeoutSync):
			for i := 1; i < int(gVar.ShardSize); i++ {
				if maskBit(i, &responsemask) == cosi.Disabled {
					it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
					SendCosiMessage(it.Address, "readyAnnoun", readyInfo{MyGlobalID, CurrentEpoch})
				}
			}
		}
	}

}

//MinerReadyProcess member use this
func MinerReadyProcess() {
	var readyMessage readyInfo
	readyMessage = <-readyCh
	for !(readyMessage.Epoch == CurrentEpoch && shard.ShardToGlobal[shard.MyMenShard.Shard][0] == readyMessage.ID) {
		readyMessage = <-readyCh
	}
	SendShardReadyMessage(LeaderAddr, "shardReady", readyInfo{MyGlobalID, CurrentEpoch})
	fmt.Println("Sent Ready")
}

func SendShardReadyMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

//HandleShardReady handle shard ready announcement command
func HandleShardReadyAnnounce(request []byte) {
	var buff bytes.Buffer
	var payload readyInfo
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	readyCh <- payload
}

//HandleShardReady handle shard ready command
func HandleShardReady(request []byte) {
	var buff bytes.Buffer
	var payload readyInfo
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	readyCh <- payload

}
