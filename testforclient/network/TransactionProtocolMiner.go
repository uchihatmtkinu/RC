package network

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/rccache"
	"github.com/uchihatmtkinu/RC/shard"
)

//HandleTx when receives a tx
func HandleTx() {
	flag := true
	sendFlag := false
	var TBCache []*basic.TransactionBatch
	for flag {
		select {
		case data := <-TxBatchCache:
			data1 := make([]byte, len(data))
			copy(data1, data)
			tmp := new(basic.TransactionBatch)
			err := tmp.Decode(&data1)
			if err == nil {
				TBCache = append(TBCache, tmp)
			}
			if !sendFlag {
				fmt.Println("Start sending packets")
				StartSendingTx <- true
				sendFlag = true
			}
		case <-time.After(timeoutGetTx):
			if len(TBCache) > 0 {
				CacheDbRef.Mu.Lock()
				tmpCnt := 0
				bad := 0
				//fmt.Println(time.Now(), "TxBatch Started", len(TBCache), "in total")
				for j := 0; j < len(TBCache); j++ {
					tmpCnt += int(TBCache[j].TxCnt)
					for i := uint32(0); i < TBCache[j].TxCnt; i++ {
						err := CacheDbRef.GetTx(&TBCache[j].TxArray[i])
						if err != nil {
							bad++
							//fmt.Println(CacheDbRef.ID, "has a error(TxBatch)", i, ": ", err)
						}
					}
				}
				//fmt.Println(time.Now(), "TxBatch Finished Total:", tmpCnt, "Bad: ", bad)
				CacheDbRef.Mu.Unlock()
				TBCache = make([]*basic.TransactionBatch, 0)
			}
		case <-StopGetTx:
			flag = false
		}
	}

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
	fmt.Println(time.Now(), CacheDbRef.ID, "gets a txlist with", tmp.TxCnt, "Txs", "Current round:", CacheDbRef.TLRound, "its round", tmp.Round, base58.Encode(tmp.HashID[:]))
	s := rccache.PreStat{Stat: -2, Valid: nil}
	CacheDbRef.Mu.Lock()
	//fmt.Println(time.Now(), "PreProcess TxList:", base58.Encode(tmp.HashID[:]))
	CacheDbRef.PreTxList(tmp, &s)
	//fmt.Println(time.Now(), "PreProcess TxList:", base58.Encode(tmp.HashID[:]), "Done")
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
	//fmt.Println(time.Now(), "Start Process TxList", base58.Encode(tmp.HashID[:]))
	CacheDbRef.Mu.Lock()
	tmpBatch := new([]basic.TransactionBatch)
	err = CacheDbRef.ProcessTL(tmp, tmpBatch)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error", err)
	}
	var sent []byte
	CacheDbRef.TLSent.Encode(&sent)
	CacheDbRef.Mu.Unlock()
	fmt.Println(time.Now(), "Start Sending TxBatch to other shards", base58.Encode(tmp.HashID[:]))
	sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxDec", sent)
	xx := shard.MyMenShard.InShardId
	BatchCache[tmp.Round] = make([]TxBatchInfo, gVar.ShardCnt)
	yy := -1
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		BatchCache[tmp.Round][i].Data = (*tmpBatch)[i].Encode()
		BatchCache[tmp.Round][i].ID = CacheDbRef.ID
		BatchCache[tmp.Round][i].ShardID = CacheDbRef.ShardNum
		BatchCache[tmp.Round][i].Round = tmp.Round
		if i != CacheDbRef.ShardNum {
			fmt.Println("Send TxBatch, Round", tmp.Round, "to", shard.ShardToGlobal[i][xx], "Shard", i)
			sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][xx]].Address, "TxMM", BatchCache[tmp.Round][i].Encode())
			if xx == int(i+1) {
				yy = int(i)
				fmt.Println("Send TxBatch, Round", tmp.Round, "to Leader", shard.ShardToGlobal[i][0], "Shard", i)
				sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][0]].Address, "TxMM", BatchCache[tmp.Round][i].Encode())
			}
		}
	}

	mask := make([]bool, gVar.ShardCnt)
	cnt = int(gVar.ShardCnt)
	if yy == -1 {
		mask[CacheDbRef.ShardNum] = true
		cnt--
	}

	for cnt > 0 {
		select {
		case nowInfo := <-txMCh[tmp.Round]:
			if shard.GlobalGroupMems[nowInfo.ID].Role == 0 {
				mask[CacheDbRef.ShardNum] = true
			} else {
				mask[shard.GlobalGroupMems[nowInfo.ID].Shard] = true
			}
			cnt--
		case <-time.After(timeoutResentTxmm):
			fmt.Println("Resend TxDec", cnt)
			for i := 0; i < len(mask); i++ {
				if !mask[i] {
					if i == int(CacheDbRef.ShardNum) {
						fmt.Println("Resend TxBatch, Round", tmp.Round, "to Leader", shard.ShardToGlobal[yy][0], "Shard", yy)
						sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[yy][0]].Address, "TxMM", BatchCache[tmp.Round][i].Encode())
					} else {
						fmt.Println("Reend TxBatch, Round", tmp.Round, "to", shard.ShardToGlobal[i][xx], "Shard", i)
						sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][xx]].Address, "TxMM", BatchCache[tmp.Round][i].Encode())
					}
				}
			}
		}
	}
	return nil
}

//HandleTxDecSet when receives a txdecset
func HandleTxDecSet(data []byte, typeInput int) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxDecSet)
	err := tmp.Decode(&data1)
	fmt.Println("Get the tds from leader:", tmp.ID, "Round:", tmp.Round)
	if typeInput == 1 {
		var tmp1 txDecRev
		tmp1.ID = CacheDbRef.ShardNum
		tmp1.Round = tmp.Round
		datax := tmp1.Encode()
		go sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxDecRev", datax)
	}
	if err != nil {
		return err
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}

	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxDecSet(tmp, &s)
	CacheDbRef.Mu.Unlock()
	if s.Stat > 0 {
		xx := shard.ShardToGlobal[tmp.ShardIndex][rand.Int()%int(gVar.ShardSize-1)+1]
		yy := txDecRev{ID: CacheDbRef.ID, Round: tmp.Round}
		go sendTxMessage(shard.GlobalGroupMems[xx].Address, "RequestTxMM", yy.Encode())
	}
	timeoutFlag := true
	cnt := s.Stat
	cntTimeout := 0
	for timeoutFlag && cnt > 0 {
		select {
		case <-s.Channel:
			cnt--
		case <-time.After(timeoutResentTxmm):
			if cntTimeout == 8 {
				fmt.Println("TDS of", tmp.ID, "Round", tmp.Round, "time out")
				timeoutFlag = false
			} else {
				xx := shard.ShardToGlobal[tmp.ShardIndex][rand.Int()%int(gVar.ShardSize-1)+1]
				yy := txDecRev{ID: CacheDbRef.ID, Round: tmp.Round}
				fmt.Println("Request TDS of", tmp.ID, "Round", tmp.Round, "from", xx)
				go sendTxMessage(shard.GlobalGroupMems[xx].Address, "RequestTxMM", yy.Encode())
				cntTimeout++
			}
		}
	}
	if tmp.Round < gVar.NumTxListPerEpoch && tmp.ShardIndex == CacheDbRef.ShardNum {
		TDSChan[tmp.Round] <- true
	}
	CacheDbRef.Mu.Lock()
	fmt.Println(time.Now(), "Miner", CacheDbRef.ID, "get TDS from", tmp.ID, "with", tmp.TxCnt, "Txs Shard", tmp.ShardIndex, "Round", tmp.Round)
	err = CacheDbRef.GetTDS(tmp)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error", err)
	}
	CacheDbRef.Mu.Unlock()
	return nil
}

//HandleAndSentTxDecSet when receives a txdecset
func HandleAndSentTxDecSet(data []byte) error {
	HandleTxDecSet(data, 1)

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
		//fmt.Println("Get txBlock from", tmp.ID, "Hash:", base58.Encode(tmp.HashID[:]), "preprocess done")
	} else {
		fmt.Println("Get txBlock from", tmp.ID, "Hash:", base58.Encode(tmp.HashID[:]), "preprocess timeout")
	}
	if tmp.Height <= CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch {
		<-TDSChan[tmp.Height-CacheDbRef.PrevHeight-1]
	}
	flag := true
	for flag {
		CacheDbRef.Mu.Lock()
		err = CacheDbRef.GetTxBlock(tmp)
		if err != nil {
			//fmt.Println("txBlock", base58.Encode(tmp.HashID[:]), " error", err)
		} else {
			flag = false
		}
		CacheDbRef.Mu.Unlock()
		time.Sleep(time.Microsecond * gVar.GeneralSleepTime)
	}

	CacheDbRef.Mu.Lock()
	fmt.Println(time.Now(), CacheDbRef.ID, "gets a txBlock with", tmp.TxCnt, "Txs from", tmp.ID, "Hash", base58.Encode(tmp.HashID[:]), "Height:", tmp.Height)
	if len(*CacheDbRef.TBCache) >= gVar.NumTxBlockForRep {
		fmt.Println(CacheDbRef.ID, "start to make repBlock")
		tmp := make([][32]byte, gVar.NumTxBlockForRep)
		copy(tmp, (*CacheDbRef.TBCache)[0:gVar.NumTxBlockForRep])
		tmpRep := shard.ReturnRepData(CacheDbRef.ShardNum)
		*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[gVar.NumTxBlockForRep:]
		startRep <- repInfo{Last: true, Hash: tmp, Rep: tmpRep}
	}
	if tmp.Height == CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch+1 {
		CacheDbRef.UnderSharding = true
		CacheDbRef.StartTxDone = false
		StopGetTx <- true
		fmt.Println(time.Now(), CacheDbRef.ID, "waits for FB")
		go WaitForFinalBlock(&shard.GlobalGroupMems)
	}
	CacheDbRef.Mu.Unlock()

	return nil
}
