package network

import (
	"fmt"
	"time"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//SendFinalBlock is to send final block
func SendFinalBlock(ms *[]shard.MemShard) {
	CacheDbRef.Mu.Lock()
	CacheDbRef.GenerateFinalBlock()

	var data []byte
	CacheDbRef.FB[CacheDbRef.ShardNum].Encode(&data, 1)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "FinalTxB", data)
		}
	}
	CacheDbRef.Mu.Unlock()
	tmp := make([][32]byte, len(*CacheDbRef.TBCache))
	copy(tmp, *CacheDbRef.TBCache)
	*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[len(*CacheDbRef.TBCache):]
	startRep <- repInfo{Last: false, Hash: tmp}
	fmt.Println(time.Now(), CacheDbRef.ID, "start to make last repBlock")
	close(StopGetTx)
}

//SendStartBlock is to send start block
func SendStartBlock(ms *[]shard.MemShard) {
	<-FinalTxReadyCh
	CacheDbRef.Mu.Lock()
	CacheDbRef.GenerateStartBlock()
	var data []byte
	CacheDbRef.TxB.Encode(&data, 1)
	fmt.Println(CacheDbRef.ID, "startBlock done")
	CacheDbRef.PrevHeight = CacheDbRef.TxB.Height
	CacheDbRef.StartTxDone = true
	CacheDbRef.Mu.Unlock()
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "StartTxB", data)
		}
	}
}

//WaitForFinalBlock is wait for final block
func WaitForFinalBlock(ms *[]shard.MemShard) error {
	data := <-finalSignal
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmpTB := new(basic.TxBlock)
	err := tmpTB.Decode(&data1, 1)
	if err != nil {
		return err
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.GetFinalTxBlock(tmpTB)
	CacheDbRef.Mu.Unlock()
	tmp := make([][32]byte, len(*CacheDbRef.TBCache))
	copy(tmp, *CacheDbRef.TBCache)
	*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[len(*CacheDbRef.TBCache):]
	startRep <- repInfo{Last: false, Hash: tmp}
	fmt.Println(time.Now(), CacheDbRef.ID, "start to make last repBlock")
	close(StopGetTx)
	return nil
}

//HandleFinalTxBlock when receives a txblock
func HandleFinalTxBlock(data []byte) error {
	finalSignal <- data
	return nil
}

//HandleStartTxBlock when receives a txblock
func HandleStartTxBlock(data []byte) error {
	<-FinalTxReadyCh
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxBlock)
	err := tmp.Decode(&data1, 1)
	if err != nil {
		return err
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.GetStartTxBlock(tmp)
	fmt.Println(CacheDbRef.ID, "startBlock done")
	CacheDbRef.StartTxDone = true
	CacheDbRef.PrevHeight = CacheDbRef.TxB.Height
	CacheDbRef.Mu.Unlock()
	return nil
}
