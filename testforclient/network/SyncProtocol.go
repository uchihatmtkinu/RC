package network

import (
	"bytes"
	"encoding/gob"
	"log"
	"sync"
	"time"

	"math/rand"

	"fmt"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
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
	SyncFlag = true
	for i := 0; i < shard.NumMems; i++ {
		//if !maskBit(i, &syncmask) {
		if i != shard.MyMenShard.Shard {
			wg.Add(1)
			go SendSyncMessage((*ms)[shard.ShardToGlobal[i][aski[i]]].Address, "requestSync", syncRequestInfo{MyGlobalID, CurrentEpoch})
			go ReceiveSyncProcess(i, &wg, ms)
		}
		//}
	}
	wg.Wait()
	SyncFlag = false
	for i := 0; i < int(gVar.ShardCnt); i++ {
		close(syncSBCh[i])
		close(syncTBCh[i])
		close(syncNotReadyCh[i])
	}

	//ShardProcess()
}

//ReceiveSyncProcess listen to the block from shard k
func ReceiveSyncProcess(k int, wg *sync.WaitGroup, ms *[]shard.MemShard) {
	fmt.Println("wait for shard", k, " from user", aski[k])
	defer wg.Done()

	//syncblock flag
	sbrxflag := true
	//txblock flag
	tbrxflag := true
	timeoutflag := false
	//txBlock Transaction block
	//TODO test
	//var txBlockMessage syncTBInfo
	//syncBlock SyncBlock
	var syncBlockMessage syncSBInfo
	for (sbrxflag || tbrxflag) && !timeoutflag {
		select {
		case syncBlockMessage = <-syncSBCh[k]:
			{
				if syncBlockMessage.Block.VerifyCoSignature(ms, k) {
					sbrxflag = false
				} else {
					aski[k] = (aski[k] + 1) % int(gVar.ShardSize)
					timeoutflag = true
					go SendSyncMessage((*ms)[shard.ShardToGlobal[k][aski[k]]].Address, "requestSync", syncRequestInfo{MyGlobalID, CurrentEpoch})
				}
			}
		//case txBlockMessage = <-syncTBCh[k]:
		//	tbrxflag = false
		case <-syncNotReadyCh[k]:
			time.Sleep(timeSyncNotReadySleep)
		case <-time.After(timeoutSync):
			{
				aski[k] = (aski[k] + 1) % int(gVar.ShardSize)
				timeoutflag = true
				go SendSyncMessage((*ms)[shard.ShardToGlobal[k][aski[k]]].Address, "requestSync", syncRequestInfo{MyGlobalID, CurrentEpoch})
				return
			}
		}
	}
	if !sbrxflag && !tbrxflag {
		//add transaction block
		//TODO test
		//CacheDbRef.GetFinalTxBlock(&txBlockMessage.Block)
		//update reputation of members
		syncBlockMessage.Block.UpdateTotalRepInMS(ms)
		//add sync Block
		Reputation.MyRepBlockChain.AddSyncBlockFromOtherShards(&syncBlockMessage.Block, k)
	} else {
		wg.Add(1)
		go ReceiveSyncProcess(k, wg, ms)
	}
	fmt.Println("received from ", k)
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
	Reputation.CurrentSyncBlock.Mu.RLock()
	if payload.Epoch > Reputation.CurrentSyncBlock.Epoch {
		SendSyncMessage(addr, "syncNReady", syncNotReadyInfo{MyGlobalID, Reputation.CurrentSyncBlock.Epoch})
		return
	}
	if payload.Epoch == Reputation.CurrentSyncBlock.Epoch {
		//TODO test
		//tmp := syncTBInfo{MyGlobalID, *(CacheDbRef.FB[CacheDbRef.ShardNum])}
		//sendTxMessage(addr, "syncTB", tmp.Encode())
		SendSyncMessage(addr, "syncSB", syncSBInfo{MyGlobalID, *Reputation.CurrentSyncBlock.Block})

	}
	Reputation.CurrentSyncBlock.Mu.RUnlock()

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

	err := payload.Decode(&request)
	if err != nil {
		log.Panic(err)
	}
	syncTBCh[shard.GlobalGroupMems[payload.ID].Shard] <- payload

}

//Encode is the encode
func (s *syncTBInfo) Encode() []byte {
	var tmp []byte
	basic.Encode(&tmp, uint32(s.ID))
	s.Block.Encode(&tmp, 1)
	return tmp
}

//Decode is the decode
func (s *syncTBInfo) Decode(buf *[]byte) error {
	var tmp int
	basic.Decode(buf, &tmp)
	s.ID = int(tmp)
	err := s.Block.Decode(buf, 1)
	return err
}
