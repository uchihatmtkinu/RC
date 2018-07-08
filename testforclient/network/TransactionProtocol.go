package network

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/rccache"

	"github.com/uchihatmtkinu/RC/gVar"

	"github.com/uchihatmtkinu/RC/shard"

	"github.com/uchihatmtkinu/RC/basic"
)

// SendTxMessage send reputation block
func SendTxMessage(addr string, command string, message []byte) {
	tmp := make([]byte, len(message))
	copy(tmp, message)
	sendTxMessage(addr, command, tmp)
}

// sendTxMessage send reputation block
func sendTxMessage(addr string, command string, message []byte) {
	request := append(commandToBytes(command), message...)
	sendData(addr, request)
}

//TxGeneralLoop is the normall loop of transaction cache
func TxGeneralLoop() {
	tmp := 0
	flag := true
	rand.Seed(time.Now().Unix())
	fmt.Println(CacheDbRef.ID, "start to process Tx")
	for {
		tmp++
		time.Sleep(time.Second * gVar.GeneralSleepTime)
		CacheDbRef.Mu.Lock()

		if CacheDbRef.Leader == CacheDbRef.ID && CacheDbRef.StartTxDone {
			if CacheDbRef.TLS == nil {
				CacheDbRef.NewTxList()
			}
			if CacheDbRef.StartIndex <= CacheDbRef.LastIndex && CacheDbRef.TDSCache[0][CacheDbRef.ShardNum].MemCnt == (gVar.ShardSize-1) {
				fmt.Println("Leader", CacheDbRef.ID, "ready to send TDS")
				CacheDbRef.SignTDS(0)
				CacheDbRef.ProcessTDS(&CacheDbRef.TDSCache[0][CacheDbRef.ShardNum])
				if flag {
					fmt.Println(CacheDbRef.ID, "sends a TxDecSet with hash:", base58.Encode(CacheDbRef.TDSCache[0][CacheDbRef.ShardNum].HashID[:]))
					data2 := new([][]byte)
					*data2 = make([][]byte, gVar.ShardCnt)

					for i := uint32(0); i < gVar.ShardCnt; i++ {
						CacheDbRef.TDSCache[0][i].Encode(&(*data2)[i])
					}
					go SendTxDecSet(*data2)
				}
				CacheDbRef.Release()
			}
			if tmp > 0 {

				//tmp = 0

				if len(CacheDbRef.Ready) >= gVar.NumTxPerBlock {

					CacheDbRef.GenerateTxBlock()
					fmt.Println(CacheDbRef.ID, "sends a TxBlock with", CacheDbRef.TxB.TxCnt, "Txs, Hash:", base58.Encode(CacheDbRef.TxB.HashID[:]))
					if len(*CacheDbRef.TBCache) >= gVar.NumTxBlockForRep {
						fmt.Println(CacheDbRef.ID, "start to make repBlock")
						tmp := make([][32]byte, gVar.NumTxBlockForRep)
						copy(tmp, (*CacheDbRef.TBCache)[0:gVar.NumTxBlockForRep])
						*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[gVar.NumTxBlockForRep:]
						startRep <- repInfo{Last: true, Hash: tmp}
					}
					data3 := new([]byte)
					CacheDbRef.TxB.Encode(data3, 0)
					go SendTxBlock(data3)
					if CacheDbRef.TxCnt >= gVar.NumTxPerEpoch {
						fmt.Println(CacheDbRef.ID, "start to make FB")
						CacheDbRef.StartTxDone = false
						CacheDbRef.StopGetTx = true
						go SendFinalBlock(&shard.GlobalGroupMems)
					}
				}
			}

			if CacheDbRef.TLS[CacheDbRef.ShardNum].TxCnt >= gVar.NumTxPerTL {
				CacheDbRef.BuildTDS()
				fmt.Println(CacheDbRef.ID, "sends a TxList with", CacheDbRef.TLS[CacheDbRef.ShardNum].TxCnt, "Txs, Hash:", base58.Encode(CacheDbRef.TLS[CacheDbRef.ShardNum].HashID[:]))
				//CacheDbRef.TLS[CacheDbRef.ShardNum].Print()
				data1 := new([]byte)
				CacheDbRef.TLS[CacheDbRef.ShardNum].Encode(data1)
				go SendTxList(*data1)
				CacheDbRef.NewTxList()
			}
		}
		CacheDbRef.Mu.Unlock()
	}
}

//TxListLoop is the loop of txlist
func TxListLoop() {

}

//SendTxList is sending txlist
func SendTxList(data []byte) {
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxList", data)
		}
	}
}

//SendTxDecSet is sending txDecSet
func SendTxDecSet(data [][]byte) {
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			//fmt.Println(CacheDbRef.ID, "send TDS to", xx)
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxDecSetM", data[CacheDbRef.ShardNum])
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		xx := rand.Int()%(int(gVar.ShardSize)-1) + 1
		if i != CacheDbRef.ShardNum {
			//fmt.Println(CacheDbRef.ID, "send TDS to", shard.ShardToGlobal[i][xx])
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

//HandleTotalTx process the tx
func HandleTotalTx(data []byte) error {
	flag := true
	for flag {
		CacheDbRef.Mu.Lock()
		if !CacheDbRef.StopGetTx {
			flag = false
		}
		CacheDbRef.Mu.Unlock()
		time.Sleep(time.Second * gVar.GeneralSleepTime)
	}

	if shard.GlobalGroupMems[CacheDbRef.ID].Role == 0 {
		HandleTxLeader(data)
	} else {
		HandleTx(data)
	}
	return nil
}

/************************Miner***************************/

//HandleTx when receives a tx
func HandleTx(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TransactionBatch)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	fmt.Println(CacheDbRef.ID, "gets a txBatch with", tmp.TxCnt, "Txs")
	CacheDbRef.Mu.Lock()
	for i := uint32(0); i < tmp.TxCnt; i++ {
		err = CacheDbRef.GetTx(&tmp.TxArray[i])
		if err != nil {
			//fmt.Println(CacheDbRef.ID, "has a error", i, ": ", err)
		}
	}
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleAndSendTx when receives a tx
func HandleAndSendTx(data []byte) error {
	HandleTotalTx(data)
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
	//fmt.Println(CacheDbRef.ID, "get TxList from", tmp.ID)
	//fmt.Println("StropGetTx", CacheDbRef.StopGetTx, "TLRound:", CacheDbRef.TLRound, "tmpRound:", tmp.Round)
	s := rccache.PreStat{Stat: -2, Valid: nil}
	flag := true
	for flag {
		CacheDbRef.Mu.Lock()
		if CacheDbRef.TLRound == tmp.Round && !CacheDbRef.StopGetTx {
			flag = false
		}
		CacheDbRef.Mu.Unlock()
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxList(tmp, &s)
	CacheDbRef.Mu.Unlock()
	for true {
		time.Sleep(time.Second * gVar.GeneralSleepTime)
		CacheDbRef.Mu.RLock()
		if s.Stat == 0 {
			CacheDbRef.Mu.RUnlock()
			break
		}
		CacheDbRef.Mu.RUnlock()
	}
	CacheDbRef.Mu.Lock()
	err = CacheDbRef.ProcessTL(tmp)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error", err)
	}
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
		time.Sleep(time.Second * gVar.GeneralSleepTime)
		CacheDbRef.Mu.RLock()
		if s.Stat == 0 {
			flag = false
		}
		CacheDbRef.Mu.RUnlock()
	}
	CacheDbRef.Mu.Lock()
	err = CacheDbRef.GetTDS(tmp)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error", err)
	}
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
	flag := true
	for flag {
		CacheDbRef.Mu.RLock()
		if CacheDbRef.StartTxDone {
			flag = false
		}
		CacheDbRef.Mu.RUnlock()
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}
	flag = true
	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxBlock(tmp, &s)
	if s.Stat == 0 {
		flag = false
	}
	CacheDbRef.Mu.Unlock()
	for flag {
		time.Sleep(time.Second * gVar.GeneralSleepTime)
		CacheDbRef.Mu.RLock()
		if s.Stat == 0 {
			flag = false
		}
		CacheDbRef.Mu.RUnlock()
	}

	flag = true
	for flag {
		CacheDbRef.Mu.Lock()
		err = CacheDbRef.GetTxBlock(tmp)
		if err != nil {
			fmt.Println(err)
		} else {
			flag = false
		}
		CacheDbRef.Mu.Unlock()
		time.Sleep(time.Microsecond * 100)
	}
	CacheDbRef.Mu.Lock()
	fmt.Println(CacheDbRef.ID, "gets a txBlock with", tmp.TxCnt, "Txs from", tmp.ID)
	if len(*CacheDbRef.TBCache) >= gVar.NumTxBlockForRep {
		fmt.Println(CacheDbRef.ID, "start to make repBlock")
		tmp := make([][32]byte, gVar.NumTxBlockForRep)
		copy(tmp, (*CacheDbRef.TBCache)[0:gVar.NumTxBlockForRep])
		*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[gVar.NumTxBlockForRep:]
		startRep <- repInfo{Last: true, Hash: tmp}
	}
	if CacheDbRef.TxCnt >= gVar.NumTxPerEpoch {
		CacheDbRef.UnderSharding = true
		CacheDbRef.StartTxDone = false
		CacheDbRef.StopGetTx = true

		fmt.Println(CacheDbRef.ID, "waits for FB")
		go WaitForFinalBlock(&shard.GlobalGroupMems)
	}
	CacheDbRef.Mu.Unlock()
	return nil
}

/*************************Leader**************************/

//HandleTxLeader when receives a tx
func HandleTxLeader(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TransactionBatch)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	fmt.Println(CacheDbRef.ID, "(Leader) gets a txBatch with", tmp.TxCnt, "Txs")
	CacheDbRef.Mu.Lock()
	for i := uint32(0); i < tmp.TxCnt; i++ {
		err = CacheDbRef.MakeTXList(&tmp.TxArray[i])
		if err != nil {
			//fmt.Println(CacheDbRef.ID, "has a error(TxBatch)", i, ": ", err)
		}
	}
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
		fmt.Println(CacheDbRef.ID, "has a error(TxDec)", err)
		return err
	}

	CacheDbRef.Mu.Lock()
	err = CacheDbRef.PreTxDecision(tmp, tmp.HashID)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error(TxDec)", err)
	}
	//tmp.Print()
	err = CacheDbRef.UpdateTXCache(tmp)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error(TxDec)", err)
	}
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
		time.Sleep(time.Second * gVar.GeneralSleepTime)
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

/*--------------Client------------*/

//HandleRequestTxB query the TxBlock
func HandleRequestTxB(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(TxBRequestInfo)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	txBs := CacheDbRef.DB.RecentBlock(uint32(tmp.Height))
	data2 := make([]byte, 0)
	basic.Encode(&data2, len(*txBs))
	for i := len(*txBs) - 1; i >= 0; i-- {
		data2 = append(data2, (*txBs)[i].Serial()...)
	}
	sendTxMessage(tmp.Address, "TxBs", data2)
	return nil
}

//Encode is encode
func (a *TxBRequestInfo) Encode() []byte {
	tmp := make([]byte, 0, 12+len(a.Address))
	basic.Encode(&tmp, []byte(a.Address))
	basic.Encode(&tmp, a.Height)
	basic.Encode(&tmp, a.Shard)
	return tmp
}

//Decode is encode
func (a *TxBRequestInfo) Decode(buf *[]byte) error {
	var xxx []byte
	err := basic.Decode(buf, &xxx)
	if err != nil {
		return err
	}
	a.Address = string(xxx)
	err = basic.Decode(buf, &a.Height)
	if err != nil {
		return err
	}
	err = basic.Decode(buf, &a.Shard)
	if err != nil {
		return err
	}
	return nil
}
