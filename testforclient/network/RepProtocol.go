package network

import (
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/gVar"
	"bytes"
	"encoding/gob"
	"log"
)
//var currentTxList *[]basic.TxList
//var currentTxDecSet *[]basic.TxDecSet

func RepProcess(ms *[]shard.MemShard) {
	var it *shard.MemShard
	flag := true
	Reputation.RepPowTxCh = make(chan Reputation.RepBlock)
	Reputation.RepPowRxCh = make(chan Reputation.RepBlock, bufferSize)
	Reputation.RepPowRxValidate = make(chan bool)
	go Reputation.MyRepBlockChain.MineRepBlock(ms, &CacheDbRef)

	for flag {
		select {
			case item := <-Reputation.RepPowTxCh: {
				for i := uint32(0); i < gVar.ShardSize; i++ {
					it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
					sendRepPowMessage(it.Address, "RepPowAnnou", item.Serialize())
				}
				flag = false
			}
			case repPowRxInfo:= <- repPowRxCh: {
				Reputation.RepPowRxCh <- handleRepPowRx(repPowRxInfo)
				if <-Reputation.RepPowRxValidate {
					flag = false
				}
			}
		}
	}
	close(Reputation.RepPowRxValidate)
	close(Reputation.RepPowTxCh)
	close(Reputation.RepPowRxCh)
}

func sendRepPowMessage(addr string, command string, message []byte ) {
	request := append(commandToBytes(command), message...)
	sendData(addr, request)
}

func handleRepPowRx(request []byte) Reputation.RepBlock {
	var buff bytes.Buffer
	var payload Reputation.RepBlock

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	return payload
}
