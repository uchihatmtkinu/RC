package network



import (
	"bytes"
	"encoding/gob"
	"log"
	"sync"
	"time"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
	"math/rand"
)

//sbrxCounter count for the block receive
//var sbrxCounter safeCounter

//aski ask user i for sync, -1 means done, otherwise ask i+1
var aski []int

//syncProcess processing sync after one epoch
func syncProcess(ms *[]shard.MemShard) {
	//waitgroup for all goroutines done
	var wg sync.WaitGroup
	aski = make([]int, int(gVar.ShardCnt))
	rand.Seed(int64(shard.MyMenShard.Shard * 3000 + shard.MyMenShard.InShardId) + time.Now().UTC().UnixNano())
	shard.PreviousSyncBlockHash = nil
	shard.PreviousSyncBlockHash = make([][32]byte, gVar.ShardCnt)
	//intilizeMaskBit(&syncmask, int((gVar.ShardCnt+7)>>3),false)
	for i := 0; i < int(gVar.ShardCnt); i++ {
		syncSBCh[i] = make(chan Reputation.SyncBlock, 10)
		syncTBCh[i] = make(chan basic.TxBlock, 10)
		syncNotReadyCh[i] = make(chan bool, 10)
		aski[i] = rand.Intn(int(gVar.ShardSize))
	}
	//TODO
	syncReady = true
	for i := 0; i < shard.NumMems; i++ {
		//if !maskBit(i, &syncmask) {
		wg.Add(1)
		go SendSyncMessage((*ms)[shard.ShardToGlobal[i][aski[i]]].Address, "requestSync", nil)
		go receiveSync(i, &wg, ms)
		//}
	}
	wg.Wait()

	for i := 0; i < int(gVar.ShardCnt); i++ {
		close(syncSBCh[i])
		close(syncTBCh[i])
	}
	shardProcess()
}

//receiveSync listen to the block from shard k
func receiveSync(k int, wg *sync.WaitGroup, ms *[]shard.MemShard) {
	defer wg.Done()
	//syncblock flag
	sbrxflag := true
	//txblock flag
	tbrxflag := true
	timeoutflag := false
	//txBlock Transaction block
	var txBlock basic.TxBlock
	//syncBlock SyncBlock
	var syncBlock Reputation.SyncBlock
	for (sbrxflag || tbrxflag) && !timeoutflag {
		select {
		case syncBlock = <-syncSBCh[k]:
			{
				if syncBlock.VerifyCoSignature(ms, k) {
					sbrxflag = false
				}	else {
					aski[k] = (aski[k] + 1) % int(gVar.ShardSize)
					timeoutflag = true
					go SendSyncMessage((*ms)[shard.ShardToGlobal[k][aski[k]]].Address, "requestSync", nil)
				}
			}
		case txBlock = <-syncTBCh[k]:
			tbrxflag = false
		case <- syncNotReadyCh[k]:
			time.Sleep(timeSyncNotReadySleep)
		case time.After(timeoutSync):
			{
				aski[k]++
				timeoutflag = true
				go SendSyncMessage((*ms)[shard.ShardToGlobal[k][aski[k]]].Address, "requestSync", nil)
			}
		}
	}
	if !sbrxflag && !tbrxflag {
		//add transaction block
		CacheDbRef.GetFinalTxBlock(&txBlock)
		//update reputation of members
		syncBlock.UpdateTotalRepInMS(ms)
		//add sync Block
		Reputation.MyRepBlockChain.AddSyncBlockFromOtherShards(&syncBlock, k)
	} else {
		wg.Add(1)
		go receiveSync(k, wg, ms)
	}
}

// sendCosiMessage send cosi message
func SendSyncMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}
// HandleChallenge rx challenge
func HandleSyncNotReady(addr string)  {
	syncNotReadyCh[GlobalAddrMapToInd[addr]] <- true
}

// HandleChallenge rx challenge
func HandleSyncSBMessage(addr string, request []byte)  {
	var buff bytes.Buffer
	var payload Reputation.SyncBlock

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	syncSBCh[GlobalAddrMapToInd[addr]] <- payload
}

//TODO
func HandleSyncTBMessage(addr string, request []byte) {
	var buff bytes.Buffer
	var payload basic.TxBlock

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	syncTBCh[GlobalAddrMapToInd[addr]] <- payload

}
