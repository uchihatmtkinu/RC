package network

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"fmt"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//var currentTxList *[]basic.TxList
//var currentTxDecSet *[]basic.TxDecSet

//RepProcessLoop is the loop of reputation
func RepProcessLoop(ms *[]shard.MemShard, NumEpoch int) {
	for i := 0; i < NumEpoch; i++ {
		flag := true
		for flag {
			flag = RepProcess(ms)
			time.Sleep(time.Second * 2)
		}
		<-startSync
	}
}

//RepProcess pow proces
func RepProcess(ms *[]shard.MemShard) bool {
	var it *shard.MemShard
	var validateFlag bool
	var item Reputation.RepPowInfo
	res := <-startRep
	flag := true
	Reputation.RepPowTxCh = make(chan Reputation.RepPowInfo)
	Reputation.RepPowRxValidate = make(chan bool)
	CacheDbRef.Mu.Lock()
	go Reputation.MyRepBlockChain.MineRepBlock(ms, &CacheDbRef)
	CacheDbRef.Mu.Unlock()
	for flag {
		select {
		case item = <-Reputation.RepPowTxCh:
			{
				for i := 0; i < int(gVar.ShardSize); i++ {
					if i != shard.MyMenShard.InShardId {
						it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
						fmt.Println("have sent " + it.Address)
						go SendRepPowMessage(it.Address, "RepPowAnnou", powInfo{MyGlobalID, item.Round, item.Hash, item.Nonce})
					}
				}
				flag = false
			}
		case validateFlag = <-Reputation.RepPowRxValidate:
			{
				if validateFlag {
					fmt.Println("add pow rep block from others")
					flag = false
				}
			}
		}
	}
	Reputation.CurrentRepBlock.Mu.RLock()
	fmt.Println("PoW ", Reputation.CurrentRepBlock.Round, " round finished.")
	Reputation.CurrentRepBlock.Mu.RUnlock()
	close(Reputation.RepPowRxValidate)
	close(Reputation.RepPowTxCh)
	return res
	//close(Reputation.RepPowRxCh)
}

// SendRepPowMessage send reputation block
func SendRepPowMessage(addr string, command string, message powInfo) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

// HandleRepPowRx receive reputation block
func HandleRepPowRx(request []byte) {
	var buff bytes.Buffer
	var payload powInfo

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	Reputation.CurrentRepBlock.Mu.RLock()
	if payload.Round > Reputation.CurrentRepBlock.Round {
		Reputation.RepPowRxCh <- Reputation.RepPowInfo{payload.Round, payload.Nonce, payload.Hash}
		fmt.Println("Received PoW from others")
	}
	Reputation.CurrentRepBlock.Mu.RUnlock()

}
