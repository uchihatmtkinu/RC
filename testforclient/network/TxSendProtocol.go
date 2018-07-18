package network

import (
	"fmt"
	"time"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//SendTx is the protocol for sending
func SendTx(x *[]byte) {
	fmt.Println(time.Now(), CacheDbRef.ID, "sending TxBatch", len(*x))
	for i := 0; i < int(gVar.ShardSize); i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxM", *x)
		}

	}

	HandleTotalTx(*x)
}

//SendLoopMiner is the protocol for sending
func SendLoopMiner(x *[]basic.Transaction) {
	fmt.Println(time.Now(), "Prepare for sending TxBatch(Miner)")
	<-StartSendingTx
	TxBatchLen := len(*x) / gVar.NumTxListPerEpoch
	for i := 0; i < gVar.NumTxListPerEpoch; i++ {
		tmp := (*x)[i*TxBatchLen : (i+1)*TxBatchLen]
		data := new(basic.TransactionBatch)
		data.New(&tmp)
		tmpHash := data.Encode()
		SendTx(&tmpHash)
		time.Sleep(time.Second * gVar.TxSendInterval)
	}
}

//SendLoopLeader is the protocol for sending
func SendLoopLeader(x *[]basic.Transaction) {
	<-StartSendTx
	fmt.Println("Prepare for sending TxBatch(Leader)")
	close(StartSendTx)
	TxBatchLen := len(*x) / gVar.NumTxListPerEpoch
	for i := 0; i < 1; i++ {
		tmp := (*x)[i*TxBatchLen : (i+1)*TxBatchLen]
		data := new(basic.TransactionBatch)
		data.New(&tmp)
		tmpHash := data.Encode()
		SendTx(&tmpHash)
		time.Sleep(time.Second * gVar.TxSendInterval)
	}
}
