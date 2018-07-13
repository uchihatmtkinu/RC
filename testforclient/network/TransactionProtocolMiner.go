package network

import (
	"fmt"
	"time"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/rccache"
	"github.com/uchihatmtkinu/RC/shard"
)

//HandleTx when receives a tx
func HandleTx(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TransactionBatch)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	//fmt.Println(time.Now(), CacheDbRef.ID, "gets a txBatch with", tmp.TxCnt, "Txs")
	flag := false

	CacheDbRef.Mu.Lock()
	if !CacheDbRef.StartSendingTX {
		flag = true
		CacheDbRef.StartSendingTX = true
	}
	for i := uint32(0); i < tmp.TxCnt; i++ {
		err = CacheDbRef.GetTx(&tmp.TxArray[i])
		if err != nil {
			//fmt.Println(CacheDbRef.ID, "has a error", i, ": ", err)
		}
	}
	CacheDbRef.Mu.Unlock()
	if flag {
		fmt.Println("Start sending packets")
		StartSendingTx <- true
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
	fmt.Println(time.Now(), CacheDbRef.ID, "gets a txlist with", tmp.TxCnt, "Txs", "Current round:", CacheDbRef.TLRound, "its round", tmp.Round)
	s := rccache.PreStat{Stat: -2, Valid: nil}
	CacheDbRef.Mu.Lock()
	fmt.Println(time.Now(), "PreProcess TxList:", base58.Encode(tmp.HashID[:]))
	CacheDbRef.PreTxList(tmp, &s)
	fmt.Println(time.Now(), "PreProcess TxList:", base58.Encode(tmp.HashID[:]), "Done")
	CacheDbRef.Mu.Unlock()
	if s.Stat != 0 {
		fmt.Println("TxList:", base58.Encode(tmp.HashID[:]), "Need waiting")
	}
	timeoutFlag := true
	cnt := s.Stat
	for timeoutFlag && cnt > 0 {
		select {
		case <-s.Channel:
			cnt--
		case <-time.After(timeoutTL):
			fmt.Println(time.Now(), "TxList:", base58.Encode(tmp.HashID[:]), "time out")
			timeoutFlag = false
		}
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
func HandleTxDecSet(data []byte, h *uint32, id *uint32) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxDecSet)
	err := tmp.Decode(&data1)
	*h = tmp.Round
	*id = tmp.ID
	if err != nil {
		return err
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}

	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxDecSet(tmp, &s)
	CacheDbRef.Mu.Unlock()

	timeoutFlag := true
	cnt := s.Stat
	for timeoutFlag && cnt > 0 {
		select {
		case <-s.Channel:
			cnt--
		case <-time.After(timeoutTL):
			fmt.Println("TDS:", base58.Encode(tmp.HashID[:]), "time out")
			timeoutFlag = false
		}
	}

	CacheDbRef.Mu.Lock()
	fmt.Println(time.Now(), "Miner", CacheDbRef.ID, "get TDS from", tmp.ID, "with", tmp.TxCnt, "Txs")
	err = CacheDbRef.GetTDS(tmp)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error", err)
	}
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleAndSentTxDecSet when receives a txdecset
func HandleAndSentTxDecSet(data []byte) error {
	var id uint32
	var round uint32
	HandleTxDecSet(data, &round, &id)
	var tmp txDecRev
	tmp.ID = CacheDbRef.ShardNum
	tmp.Round = round
	datax := tmp.Encode()
	sendTxMessage(shard.GlobalGroupMems[id].Address, "TxDecRev", datax)
	fmt.Println(CacheDbRef.ID, "Get TDS and send")
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
	fmt.Println("Get txBlock from", tmp.ID, "Hash:", base58.Encode(tmp.HashID[:]))
	if err != nil {
		return err
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}

	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxBlock(tmp, &s)
	CacheDbRef.Mu.Unlock()

	timeoutFlag := true
	cnt := s.Stat
	for timeoutFlag && cnt > 0 {
		select {
		case <-s.Channel:
			cnt--
		case <-time.After(timeoutTL):
			timeoutFlag = false
		}
	}
	if cnt == 0 {
		fmt.Println("Block", base58.Encode(tmp.HashID[:]), "preprocess done")
	} else {
		fmt.Println("Block", base58.Encode(tmp.HashID[:]), "preprocess timeout")
	}

	CacheDbRef.Mu.Lock()
	err = CacheDbRef.GetTxBlock(tmp)
	if err != nil {
		fmt.Println("txBlock", base58.Encode(tmp.HashID[:]), " error", err)
	}
	CacheDbRef.Mu.Unlock()

	CacheDbRef.Mu.Lock()
	fmt.Println(time.Now(), CacheDbRef.ID, "gets a txBlock with", tmp.TxCnt, "Txs from", tmp.ID)
	if len(*CacheDbRef.TBCache) >= gVar.NumTxBlockForRep {
		fmt.Println(CacheDbRef.ID, "start to make repBlock")
		tmp := make([][32]byte, gVar.NumTxBlockForRep)
		copy(tmp, (*CacheDbRef.TBCache)[0:gVar.NumTxBlockForRep])
		*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[gVar.NumTxBlockForRep:]
		startRep <- repInfo{Last: true, Hash: tmp}
	}
	if CacheDbRef.TxB.Height == CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch+1 {
		CacheDbRef.UnderSharding = true
		CacheDbRef.StartTxDone = false
		CacheDbRef.StopGetTx = true

		fmt.Println(CacheDbRef.ID, "waits for FB")
		go WaitForFinalBlock(&shard.GlobalGroupMems)
	}
	CacheDbRef.Mu.Unlock()
	return nil
}
