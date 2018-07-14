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
	if CurrentEpoch != -1 {
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
	fmt.Println(time.Now(), CacheDbRef.ID, "Shard Calculated")
	LeaderAddr = shard.GlobalGroupMems[shard.ShardToGlobal[shard.MyMenShard.Shard][0]].Address
	CacheDbRef.Mu.Lock()
	CacheDbRef.DB.ClearTx()
	CacheDbRef.TDSCnt = make([]int, gVar.ShardCnt)
	CacheDbRef.TDSNotReady = int(gVar.ShardCnt)
	StopGetTx = make(chan bool, 1)
	CacheDbRef.ShardNum = uint32(shard.MyMenShard.Shard)
	CacheDbRef.Leader = uint32(shard.ShardToGlobal[shard.MyMenShard.Shard][0])
	CacheDbRef.Mu.Unlock()
	TxBatchCache = make(chan []byte, 1000)
	if shard.MyMenShard.Role == 1 {
		MinerReadyProcess()
	} else {
		if CurrentEpoch != -1 {
			go SendStartBlock(&shard.GlobalGroupMems)
		}
		LeaderReadyProcess(&shard.GlobalGroupMems)
	}

	fmt.Println("shard finished")
	if CurrentEpoch != -1 {
		CacheDbRef.Clear()
		FinalTxReadyCh <- true
	}
}

//LeaderReadyProcess leader use this
func LeaderReadyProcess(ms *[]shard.MemShard) {
	var readyMessage readyInfo
	var it *shard.MemShard
	var membermask []byte
	var leadermask []byte
	intilizeMaskBit(&membermask, (int(gVar.ShardSize)+7)>>3, cosi.Disabled)
	intilizeMaskBit(&leadermask, (int(gVar.ShardCnt)+7)>>3, cosi.Disabled)
	readyMember := 1
	readyLeader := 1
	//sent announcement
	for i := 1; i < int(gVar.ShardSize); i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		SendShardReadyMessage(it.Address, "readyAnnoun", readyInfo{MyGlobalID, CurrentEpoch})
	}
	//fmt.Println("wait for ready")
	//TODO modify int(gVar.ShardSize)/2
	for readyMember < int(gVar.ShardSize) {
		select {
		case readyMessage = <-readyMemberCh:
			if readyMessage.Epoch == CurrentEpoch {
				readyMember++
				setMaskBit((*ms)[readyMessage.ID].InShardId, cosi.Enabled, &membermask)
				//fmt.Println("ReadyCount: ", readyCount)
			}
		case <-time.After(timeoutSync):
			//fmt.Println("Wait shard signal time out")
			for i := 1; i < int(gVar.ShardSize); i++ {
				if maskBit(i, &membermask) == cosi.Disabled {
					it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
					SendShardReadyMessage(it.Address, "readyAnnoun", readyInfo{MyGlobalID, CurrentEpoch})
				}
			}
		}
	}
	fmt.Println(time.Now(), "Shard is ready, sent to other shards")
	for i := 0; i < int(gVar.ShardCnt); i++ {
		if i != shard.MyMenShard.Shard {
			it = &(*ms)[shard.ShardToGlobal[i][0]]
			SendShardReadyMessage(it.Address, "leaderReady", readyInfo{shard.MyMenShard.Shard, CurrentEpoch})
		}
	}

	for readyLeader < int(gVar.ShardCnt) {
		select {
		case readyMessage = <-readyLeaderCh:
			if readyMessage.Epoch == CurrentEpoch {
				readyLeader++
				setMaskBit(readyMessage.ID, cosi.Enabled, &leadermask)
				fmt.Println(time.Now(), "ReadyLeaderCount: ", readyLeader)
			}
		case <-time.After(timeoutSync):
			for i := 1; i < int(gVar.ShardCnt); i++ {
				if maskBit(i, &leadermask) == cosi.Disabled {
					fmt.Println(time.Now(), "Send ReadyLeader to Shard", i, "ID", shard.ShardToGlobal[i][0])
					it = &(*ms)[shard.ShardToGlobal[i][0]]
					SendShardReadyMessage(it.Address, "leaderReady", readyInfo{shard.MyMenShard.Shard, CurrentEpoch})
				}
			}
		}

	}
	fmt.Println("All shards are ready.")
}

//MinerReadyProcess member use this
func MinerReadyProcess() {
	var readyMessage readyInfo
	readyMessage = <-readyMemberCh
	for !(readyMessage.Epoch == CurrentEpoch && shard.ShardToGlobal[shard.MyMenShard.Shard][0] == readyMessage.ID) {
		readyMessage = <-readyMemberCh
	}
	SendShardReadyMessage(LeaderAddr, "shardReady", readyInfo{MyGlobalID, CurrentEpoch})
	fmt.Println("Sent Ready")
}

func SendShardReadyMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
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
	readyMemberCh <- payload

}

//HandleLeaderReady handle shard ready command from other leader
func HandleLeaderReady(request []byte) {
	var buff bytes.Buffer
	var payload readyInfo
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	readyLeaderCh <- payload
}
