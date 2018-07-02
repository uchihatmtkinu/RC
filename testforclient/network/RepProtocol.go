package network

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
	"fmt"
	"sync"
)

//var currentTxList *[]basic.TxList
//var currentTxDecSet *[]basic.TxDecSet

//RepProcess pow process
func RepProcess(ms *[]shard.MemShard) {
	var it *shard.MemShard
	var validateFlag bool
	var item *Reputation.RepBlock
	var wg sync.WaitGroup
	flag := true
	Reputation.RepPowTxCh = make(chan *Reputation.RepBlock)
	Reputation.RepPowRxValidate = make(chan bool)
	go Reputation.MyRepBlockChain.MineRepBlock(ms, &CacheDbRef)

	for flag {
		select {
		case item = <-Reputation.RepPowTxCh:
			{
				Reputation.CurrentRepBlock.Mu.Lock()
				for i := 0; i < int(gVar.ShardSize); i++ {
					if i != shard.MyMenShard.InShardId {
						it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
						fmt.Println("have sent"+it.Address)
						wg.Add(1)
						go SendRepPowMessage(it.Address, "RepPowAnnou", powInfo{MyGlobalID, Reputation.CurrentRepBlock.Round, Reputation.CurrentRepBlock.Block.Hash, Reputation.CurrentRepBlock.Block.Nonce}, &wg)
					}
				}
				wg.Wait()
				Reputation.CurrentRepBlock.Mu.Unlock()
				flag = false
			}
		case validateFlag = <-Reputation.RepPowRxValidate:
			{
				if validateFlag {
					flag = false
				}
			}
		}
	}
	close(Reputation.RepPowRxValidate)
	close(Reputation.RepPowTxCh)
	//close(Reputation.RepPowRxCh)
}


// SendRepPowMessage send reputation block
func SendRepPowMessage(addr string, command string, message powInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

// HandleRepPowRx receive reputation block
func HandleRepPowRx(request []byte)  {
	var buff bytes.Buffer
	var payload powInfo

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	Reputation.CurrentRepBlock.Mu.Lock()
	defer Reputation.CurrentRepBlock.Mu.Unlock()
	if payload.Round >= Reputation.CurrentRepBlock.Round{
		Reputation.RepPowRxCh  <- Reputation.RepPowRxInfo{payload.Nonce, payload.Hash}
	}

}
