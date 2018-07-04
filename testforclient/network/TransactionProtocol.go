package network

import (
	"math/rand"
	"time"

	"github.com/uchihatmtkinu/RC/rccache"

	"github.com/uchihatmtkinu/RC/gVar"

	"github.com/uchihatmtkinu/RC/shard"

	"github.com/uchihatmtkinu/RC/basic"
)

// sendRepPowMessage send reputation block
func sendTxMessage(addr string, command string, message []byte) {
	request := append(commandToBytes(command), message...)
	sendData(addr, request)
}

//TxGeneralLoop is the normall loop of transaction cache
func TxGeneralLoop() {
	tmp := 0
	flag := false
	rand.Seed(time.Now().Unix())
	for {
		tmp++
		time.Sleep(time.Second * 10)

		CacheDbRef.Mu.Lock()
		if CacheDbRef.StartIndex <= CacheDbRef.LastIndex && CacheDbRef.TDSCache[0][0].MemCnt > (gVar.ShardSize-1)/2 {
			CacheDbRef.SignTDS(0)

			if flag {
				data2 := new([][]byte)
				*data2 = make([][]byte, gVar.ShardCnt)
				for i := uint32(0); i < gVar.ShardCnt; i++ {
					CacheDbRef.TDSCache[0][i].Encode(&(*data2)[i])
				}
				go SendTxDecSet(*data2)
			}
			CacheDbRef.Release()
			if tmp == 3 {
				flag = true
				tmp = 0
				if len(CacheDbRef.Ready) > 1 {
					CacheDbRef.GenerateTxBlock()
					data3 := new([]byte)
					CacheDbRef.TxB.Encode(data3, 0)
					go SendTxBlock(data3)
				}
			}
		}
		CacheDbRef.Mu.Unlock()

		CacheDbRef.Mu.Lock()
		if CacheDbRef.TLS[CacheDbRef.ShardNum].TxCnt >= 1 {
			CacheDbRef.BuildTDS()
			data1 := new([]byte)
			CacheDbRef.TLS[CacheDbRef.ShardNum].Encode(data1)
			go SendTxList(data1)
			CacheDbRef.NewTxList()
		}
		CacheDbRef.Mu.Unlock()
	}
}

//SendTxList is sending txlist
func SendTxList(data *[]byte) {
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxList", *data)
		}
	}
}

//SendTxDecSet is sending txDecSet
func SendTxDecSet(data [][]byte) {
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxDecSetM", data[CacheDbRef.ShardNum])
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		xx := rand.Int()%(int(gVar.ShardSize)-1) + 1
		if i != CacheDbRef.ShardNum {
			sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][xx]].Address, "TxDecSet", data[i])
		}
	}
}

//SendTxBlock is sending txBlock
func SendTxBlock(data *[]byte) {
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxB", *data)
		}
	}
}

/************************Miner***************************/

//HandleTx when receives a tx
func HandleTx(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.Transaction)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.GetTx(tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleAndSendTx when receives a tx
func HandleAndSendTx(data []byte) error {
	HandleTx(data)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxM", data)
		}
	}
	return nil
}

//HandleTxList when receives a txlist
func HandleTxList(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxList)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}
	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxList(tmp, &s)
	CacheDbRef.Mu.Unlock()
	for true {
		time.Sleep(time.Second)
		CacheDbRef.Mu.RLock()
		if s.Stat == 0 {
			CacheDbRef.Mu.RUnlock()
			break
		}
		CacheDbRef.Mu.RUnlock()
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.ProcessTL(tmp)
	var sent []byte
	CacheDbRef.TLSent.Encode(&sent)
	CacheDbRef.Mu.Unlock()
	sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxDec", sent)
	return nil
}

//HandleTxDecSet when receives a txdecset
func HandleTxDecSet(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxDecSet)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}
	flag := true
	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxDecSet(tmp, &s)
	if s.Stat == 0 {
		flag = false
	}
	CacheDbRef.Mu.Unlock()
	for flag {
		time.Sleep(time.Second)
		CacheDbRef.Mu.RLock()
		if s.Stat == 0 {
			flag = false
		}
		CacheDbRef.Mu.RUnlock()
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.GetTDS(tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleAndSentTxDecSet when receives a txdecset
func HandleAndSentTxDecSet(data []byte) error {
	HandleTxDecSet(data)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxDecSetM", data)
		}
	}

	return nil
}

//HandleTxBlock when receives a txblock
func HandleTxBlock(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxBlock)
	err := tmp.Decode(&data1, 0)
	if err != nil {
		return err
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}
	flag := true
	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxBlock(tmp, &s)
	if s.Stat == 0 {
		flag = false
	}
	CacheDbRef.Mu.Unlock()
	for flag {
		time.Sleep(time.Second)
		CacheDbRef.Mu.RLock()
		if s.Stat == 0 {
			flag = false
		}
		CacheDbRef.Mu.RUnlock()
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.GetTxBlock(tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleFinalTxBlock when receives a txblock
func HandleFinalTxBlock(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxBlock)
	err := tmp.Decode(&data1, 1)
	if err != nil {
		return err
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.GetFinalTxBlock(tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleAndSentFinalTxBlock when receives a txblock
func HandleAndSentFinalTxBlock(data []byte) error {
	HandleFinalTxBlock(data)
	xx := shard.MyMenShard.InShardId
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if i != CacheDbRef.ShardNum {
			sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][xx]].Address, "FinalTxBM", data)
		}
	}
	return nil
}

/*************************Leader**************************/

//HandleTxLeader when receives a tx
func HandleTxLeader(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.Transaction)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.MakeTXList(tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleTxDecLeader when receives a txdec
func HandleTxDecLeader(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxDecision)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}

	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxDecision(tmp, tmp.HashID)
	CacheDbRef.UpdateTXCache(tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleTxDecSetLeader when receives a txdecset
func HandleTxDecSetLeader(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxDecSet)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}
	flag := true
	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxDecSet(tmp, &s)
	if s.Stat == 0 {
		flag = false
	}
	CacheDbRef.Mu.Unlock()
	for flag {
		time.Sleep(time.Second)
		CacheDbRef.Mu.Lock()
		if s.Stat == 0 {
			flag = false
		}
		CacheDbRef.Mu.Unlock()
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.ProcessTDS(tmp)
	CacheDbRef.Mu.Unlock()
	return nil
}
