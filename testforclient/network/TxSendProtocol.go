package network

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//SendTx is the protocol for sending
func SendTx(x *[]byte) {
	fmt.Println(time.Now(), CacheDbRef.ID, "sending TxBatch")
	for i := 0; i < int(gVar.ShardSize); i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxM", *x)
		}

	}
	for i := 0; i < int(gVar.ShardCnt); i++ {
		xx := rand.Int()%(int(gVar.ShardSize)-1) + 1
		if i != int(CacheDbRef.ShardNum) {
			sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][xx]].Address, "Tx", *x)
		}
	}
	HandleTotalTx(*x)
}

//SendLoop is the protocol for sending
func SendLoop(x *[]basic.Transaction) {

	for i := 0; i < len(*x)/gVar.NumTxPerTL; i++ {
		tmp := (*x)[i*gVar.NumTxPerTL : (i+1)*gVar.NumTxPerTL]
		data := new(basic.TransactionBatch)
		data.New(&tmp)
		tmpHash := data.Encode()
		SendTx(&tmpHash)
		time.Sleep(time.Second * gVar.TxSendInterval)
	}
}
