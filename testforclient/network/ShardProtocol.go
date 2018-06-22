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
)

//sbrxCounter count for the block receive
var sbrxCounter safeCounter

//aski ask user i for sync, -1 means done, otherwise ask i+1
var aski []int

//syncProcess processing sync after one epoch
func syncProcess() {
	//waitgroup for all goroutines done
	var wg sync.WaitGroup
	aski = make([]int, int(gVar.ShardCnt))
	//intilizeMaskBit(&syncmask, int((gVar.ShardCnt+7)>>3),false)
	for i := 0; i < int(gVar.ShardCnt); i++ {
		syncSBCh[i] = make(chan sbInfoCh, 10)
		syncTBCh[i] = make(chan sbInfoCh, 10)
		aski[i] = 0
	}
	for i := 0; i < shard.NumMems; i++ {
		//if !maskBit(i, &syncmask) {
		wg.Add(1)
		go sendSyncMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][aski[i]]].Address, "requestsync", nil)
		go receiveSync(i, &wg)
		//}
	}
	wg.Wait()

}

//receiveSync listen to the block from shard k
func receiveSync(k int, wg *sync.WaitGroup) {
	defer wg.Done()
	//syncblock flag
	sbrxflag := true
	//txblock flag
	tbrxflag := true
	timeoutflag := false
	for (sbrxflag || tbrxflag) && !timeoutflag {
		select {
		case sbinfo := <-syncSBCh[k]:
			{
				if sbinfo.id == aski[k] {
					syncBlock := handleSBMessage(sbinfo.request)
					sbrxflag = false
					//TODO add SyncBlock
				}
				//sbRxFlag[k] <- true

			}
		case tbinfo := <-syncTBCh[k]:
			{
				if tbinfo.id == aski[k] {
					txBlock := handleTBMessage(tbinfo.request)
					tbrxflag = false
					CacheDbRef.GetFinalTxBlock(&txBlock)
					//TODO add TxBlock
				}
				//tbRxFlag[k] <- true

			}
		case time.After(timeoutSync):
			{
				aski[k]++
				timeoutflag = true
				go sendSyncMessage(shard.GlobalGroupMems[shard.ShardToGlobal[k][aski[k]]].Address, "requestsync", nil)
			}
		}
	}
	if !sbrxflag && !tbrxflag {
		sbrxCounter.mux.Lock()
		sbrxCounter.cnt++
		sbrxCounter.mux.Unlock()
	} else {
		wg.Add(1)
		go receiveSync(k, wg)
	}
}

// sendCosiMessage send cosi message
func sendSyncMessage(addr string, command string, message interface{}) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

// handleChallenge rx challenge
func handleSBMessage(request []byte) Reputation.SyncBlock {
	var buff bytes.Buffer
	var payload Reputation.SyncBlock

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	return payload
}

func handleTBMessage(request []byte) basic.TxBlock {
	var buff bytes.Buffer
	var payload basic.TxBlock

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	return payload
}
