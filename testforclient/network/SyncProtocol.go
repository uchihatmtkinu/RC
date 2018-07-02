package network

import (
	"bytes"
	"encoding/gob"
	"log"
	"sync"
	"time"

	"math/rand"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
	"fmt"
)

//sbrxCounter count for the block receive
//var sbrxCounter safeCounter

//aski ask user i for sync, -1 means done, otherwise ask i+1
var aski []int

//SyncProcess processing sync after one epoch
func SyncProcess(ms *[]shard.MemShard) {
	CurrentEpoch++
	fmt.Println("Sync Began")


	//waitgroup for all goroutines done
	var wg sync.WaitGroup
	aski = make([]int, int(gVar.ShardCnt))
	rand.Seed(int64(shard.MyMenShard.Shard*3000+shard.MyMenShard.InShardId) + time.Now().UTC().UnixNano())
	shard.PreviousSyncBlockHash = make([][32]byte, gVar.ShardCnt)
	//intilizeMaskBit(&syncmask, int((gVar.ShardCnt+7)>>3),false)
	for i := 0; i < int(gVar.ShardCnt); i++ {
		syncSBCh[i] = make(chan syncSBInfo, 10)
		syncTBCh[i] = make(chan syncTBInfo, 10)
		syncNotReadyCh[i] = make(chan bool, 10)
		aski[i] = rand.Intn(int(gVar.ShardSize))
	}
	for i := 0; i < shard.NumMems; i++ {
		//if !maskBit(i, &syncmask) {
		if i!=shard.MyMenShard.Shard {
			wg.Add(1)
			go SendSyncMessage((*ms)[shard.ShardToGlobal[i][aski[i]]].Address, "requestSync", syncRequestInfo{MyGlobalID, CurrentEpoch})
			go receiveSync(i, &wg, ms)
		}
		//}
	}
	wg.Wait()

	for i := 0; i < int(gVar.ShardCnt); i++ {
		close(syncSBCh[i])
		close(syncTBCh[i])
		close(syncNotReadyCh[i])
	}
	ShardProcess()
}

//receiveSync listen to the block from shard k
func receiveSync(k int, wg *sync.WaitGroup, ms *[]shard.MemShard) {
	fmt.Println("wait for shard", k, " from user", aski)
	defer wg.Done()
	//syncblock flag
	sbrxflag := true
	//txblock flag
	tbrxflag := false
	timeoutflag := false
	//txBlock Transaction block
	//var txBlock basic.TxBlock
	//syncBlock SyncBlock
	var syncBlock syncSBInfo
	for (sbrxflag || tbrxflag) && !timeoutflag {
		select {
		case syncBlock = <-syncSBCh[k]:
			{
				if syncBlock.Block.VerifyCoSignature(ms, k) {
					sbrxflag = false
				} else {
					aski[k] = (aski[k] + 1) % int(gVar.ShardSize)
					timeoutflag = true
					go SendSyncMessage((*ms)[shard.ShardToGlobal[k][aski[k]]].Address, "requestSync", syncRequestInfo{MyGlobalID, CurrentEpoch})
				}
			}
		//case txBlock = <-syncTBCh[k]:
		//	tbrxflag = false
		case <-syncNotReadyCh[k]:
			time.Sleep(timeSyncNotReadySleep)
		case <-time.After(timeoutSync):
			{
				aski[k] = (aski[k] + 1) % int(gVar.ShardSize)
				timeoutflag = true
				go SendSyncMessage((*ms)[shard.ShardToGlobal[k][aski[k]]].Address, "requestSync", syncRequestInfo{MyGlobalID, CurrentEpoch} )
				return
			}
		}
	}
	if !sbrxflag && !tbrxflag {
		//add transaction block
		//CacheDbRef.GetFinalTxBlock(&txBlock)
		//update reputation of members
		syncBlock.Block.UpdateTotalRepInMS(ms)
		//add sync Block
		Reputation.MyRepBlockChain.AddSyncBlockFromOtherShards(&syncBlock.Block, k)
	} else {
		wg.Add(1)
		go receiveSync(k, wg, ms)
	}
	fmt.Println("received from ",k)
}

// sendCosiMessage send cosi message
func SendSyncMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

// HandleChallenge rx challenge
func HandleSyncNotReady(request []byte) {
	var buff bytes.Buffer
	var payload syncNotReadyInfo
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	syncNotReadyCh[shard.GlobalGroupMems[payload.ID].Shard] <- true
}
func HandleRequestSync(request []byte) {
	var buff bytes.Buffer
	var payload syncRequestInfo

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	addr := shard.GlobalGroupMems[payload.ID].Address
	if payload.Epoch > CurrentEpoch {
		go SendSyncMessage(addr, "syncNReady", "syncNReady")
	} else {
		go sendTxMessage(addr, "syncTB", syncTBInfo{MyGlobalID, CacheDbRef.FB[CacheDbRef.ShardNum].Serial()})
		go SendSyncMessage(addr, "syncSB", syncSBInfo{MyGlobalID, *Reputation.CurrentSyncBlock})
	}

}

// HandleChallenge rx challenge
func HandleSyncSBMessage(request []byte) {
	var buff bytes.Buffer
	var payload syncSBInfo

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	syncSBCh[shard.GlobalGroupMems[payload.ID].Shard] <- payload
}

//HandleSyncTBMessage decodes the txblock
func HandleSyncTBMessage(request []byte) {
	var payload syncTBInfo

	err := payload.Decode(&request, 1)
	if err != nil {
		log.Panic(err)
	}
	syncTBCh[shard.GlobalGroupMems[payload.ID].Shard] <- payload

}
