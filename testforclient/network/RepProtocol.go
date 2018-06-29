package network

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//var currentTxList *[]basic.TxList
//var currentTxDecSet *[]basic.TxDecSet

//RepProcess pow process
func RepProcess(ms *[]shard.MemShard) {
	var it *shard.MemShard
	var validateFlag bool
	flag := true
	Reputation.RepPowTxCh = make(chan Reputation.RepBlock)
	Reputation.RepPowRxCh = make(chan Reputation.RepBlock, bufferSize)
	Reputation.RepPowRxValidate = make(chan bool)
	go Reputation.MyRepBlockChain.MineRepBlock(ms, &CacheDbRef)

	for flag {
		select {
		case item := <-Reputation.RepPowTxCh:
			{
				for i := 0; i < int(gVar.ShardSize); i++ {
					if i != shard.MyMenShard.Shard {
						it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
						SendRepPowMessage(it.Address, "RepPowAnnou", item.Serialize())
					}
				}
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
	close(Reputation.RepPowRxCh)
}


// SendRepPowMessage send reputation block
func SendRepPowMessage(addr string, command string, message []byte) {
	request := append(commandToBytes(command), message...)
	sendData(addr, request)
}

// HandleRepPowRx receive reputation block
func HandleRepPowRx(request []byte)  {
	var buff bytes.Buffer
	var payload Reputation.RepBlock

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	Reputation.RepPowRxCh  <- payload
}
