package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//MakeTXList is to create TxList given transaction
func (d *DbRef) MakeTXList(b *basic.Transaction) error {
	tmpPre, okw := d.WaitHashCache[basic.HashCut(b.Hash)]
	if okw {
		for i := 0; i < len(tmpPre.DataTB); i++ {
			tmpPre.StatTB[i].Valid[tmpPre.IDTB[i]] = 1
			tmpPre.StatTB[i].Stat--
			tmpPre.DataTB[i].TxArray[tmpPre.IDTB[i]] = *b
		}
		for i := 0; i < len(tmpPre.DataTDS); i++ {
			tmpPre.StatTDS[i].Valid[tmpPre.IDTDS[i]] = 1
			tmpPre.StatTDS[i].Stat--
			tmpPre.DataTDS[i].TxArray[tmpPre.IDTDS[i]] = b.Hash
		}
		for i := 0; i < len(tmpPre.DataTL); i++ {
			tmpPre.StatTL[i].Valid[tmpPre.IDTL[i]] = 1
			tmpPre.StatTL[i].Stat--
			tmpPre.DataTL[i].TxArray[tmpPre.IDTL[i]] = b.Hash
		}
	}
	delete(d.WaitHashCache, basic.HashCut(b.Hash))
	tmp, ok := d.TXCache[b.Hash]
	if !ok {
		tmp = new(CrossShardDec)
		tmp.New(b)
	} else {
		tmp = new(CrossShardDec)
		tmp.Update(b)
	}
	if tmp.InCheck[d.ShardNum] == 0 {
		if ok {
			delete(d.TXCache, b.Hash)
		}
		return fmt.Errorf("Not related TX")
	}
	if tmp.Res == 1 {
		d.Ready = append(d.Ready, *tmp.Data)
	}
	d.AddCache(b.Hash)
	d.TXCache[b.Hash] = tmp
	if d.TLS == nil {
		d.NewTxList()
	}
	if tmp.InCheck[d.ShardNum] != -1 {
		for i := uint32(0); i < gVar.ShardCnt; i++ {
			if tmp.InCheck[i] != 0 {
				d.TLS[i].AddTx(b)
			}
		}
	}
	return nil
}

//BuildTDS is to build all txDecSet
//Must after SignTXL
func (d *DbRef) BuildTDS() {
	d.TLS[d.ShardNum].Sign(&d.prk)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if d.ShardNum != i {
			d.TLS[i].HashID = d.TLS[i].Hash()
		}
	}
	d.TDS = new([gVar.ShardCnt]basic.TxDecSet)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if i == d.ShardNum {
			d.TDS[i].Set(&d.TLS[i], d.ShardNum, 1)
		} else {
			d.TDS[i].Set(&d.TLS[i], d.ShardNum, 0)
		}
	}

}

//SignTDS is to sign all txDecSet
func (d *DbRef) SignTDS(x int) {
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.TDSCache[x][i].Sign(&d.prk)
	}

}

//NewTxList initialize the txList
//Must after BuildTDS
func (d *DbRef) NewTxList() error {
	if d.TLS != nil {
		//d.TLCache = append(d.TLCache, *d.TL)
		d.TLSCache = append(d.TLSCache, *d.TLS)
		d.TDSCache = append(d.TDSCache, *d.TDS)
		d.LastIndex++
		d.TLRound++
		d.TLIndex[d.TLS[d.ShardNum].Hash()] = uint32(d.LastIndex)
	}

	d.TLS = new([gVar.ShardCnt]basic.TxList)

	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.TLS[i].ID = d.ID
		d.TLS[i].Round = d.TLRound
	}
	if d.TLRound == gVar.NumTxListPerEpoch {
		d.StopGetTx = true
	}
	//d.TL = new(basic.TxList)
	//d.TL.Set(d.ID)
	return nil
}

//GenerateTxBlock makes the TxBlock
func (d *DbRef) GenerateTxBlock() error {
	height := d.TxB.Height
	d.TxB = new(basic.TxBlock)
	d.TxB.MakeTxBlock(d.ID, &d.Ready, d.DB.LastTB, &d.prk, height+1, 0, nil, 0)

	for i := 0; i < len(d.Ready); i++ {
		d.ClearCache(d.Ready[i].Hash)
	}
	d.Ready = nil
	*(d.TBCache) = append(*(d.TBCache), d.TxB.HashID)

	d.DB.AddBlock(d.TxB)
	d.DB.UpdateUTXO(d.TxB, d.ShardNum)
	d.DB.ShowAccount()

	return nil
}

//GenerateFinalBlock generate final block
func (d *DbRef) GenerateFinalBlock() error {
	tmp := d.DB.MakeFinalTx(d.ShardNum)
	height := d.TxB.Height
	d.TxB = new(basic.TxBlock)
	d.TxB.MakeTxBlock(d.ID, tmp, d.DB.LastFB[d.ShardNum], &d.prk, height+1, 1, &d.FB[d.ShardNum].HashID, d.ShardNum)
	*(d.TBCache) = append(*(d.TBCache), d.TxB.HashID)
	d.FB[d.ShardNum] = d.TxB
	err := d.DB.AddFinalBlock(d.TxB)
	if err != nil {
		fmt.Println(err)
	}
	xxx := d.TxB.Serial()
	//fmt.Println("FB length: ", len(xxx))
	xxxtmp := new(basic.TxBlock)
	err = xxxtmp.Decode(&xxx, 1)
	//fmt.Println(err)
	return nil
}

//GenerateStartBlock generate Start block
func (d *DbRef) GenerateStartBlock() error {
	height := d.FB[d.ShardNum].Height
	tmp := make([][32]byte, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		tmp[i] = d.FB[i].HashID
	}
	d.TxB = new(basic.TxBlock)
	d.TxB.MakeStartBlock(d.ID, &tmp, d.DB.LastFB[d.ShardNum], &d.prk, height+1)
	d.DB.AddStartBlock(d.TxB)
	return nil
}

//UpdateTXCache is to pick the transactions into ready slice given txdecision
func (d *DbRef) UpdateTXCache(a *basic.TxDecision) error {
	if a.Single == 1 {
		return fmt.Errorf("TxDecision parameter error")
	}
	tmp, ok := d.TLIndex[a.HashID]
	if !ok {
		return fmt.Errorf("TxDecision Hash error, wrong or time out")
	}
	tmpIndex := tmp - uint32(d.StartIndex)
	tmpTL := d.TLSCache[tmpIndex][d.ShardNum]

	var x, y uint32 = 0, 0
	tmpTD := make([]basic.TxDecision, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {

		tmpTD[i].Set(a.ID, i, 1)
		tmpTD[i].HashID = d.TLSCache[tmpIndex][i].HashID
		tmpTD[i].Sig = nil
		tmpTD[i].Sig = append(tmpTD[i].Sig, a.Sig[i])
	}
	//fmt.Println("Leader ", d.ID, " process TxDecision: ")
	for i := uint32(0); i < tmpTL.TxCnt; i++ {
		tmpTx := d.TXCache[tmpTL.TxArray[i]]
		//tmpTx.Print()
		for j := 0; j < len(tmpTx.ShardRelated); j++ {
			tmpTD[tmpTx.ShardRelated[j]].Add((a.Decision[x] >> y) & 1)
		}
		if y < 7 {
			y++
		} else {
			x++
			y = 0
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		//fmt.Println("Decision: ", tmpTD[i].Decision, "Sig: ", tmpTD[i].Sig[0])
		if !tmpTD[i].Verify(&shard.GlobalGroupMems[a.ID].RealAccount.Puk, 0) {
			return fmt.Errorf("Signature not match %d", i)
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.TDSCache[tmpIndex][i].Add(&tmpTD[i])
	}
	return nil
}

//ProcessTDS deal with the TDS
func (d *DbRef) ProcessTDS(b *basic.TxDecSet) {
	if b.ShardIndex == d.ShardNum {
		tmp, _ := d.TLIndex[b.HashID]
		tmpIndex := tmp - uint32(d.StartIndex)
		tmpTL := d.TLSCache[tmpIndex][d.ShardNum]
		b.TxCnt = tmpTL.TxCnt
		b.TxArray = tmpTL.TxArray
		index := 0
		shift := byte(0)
		for i := uint32(0); i < b.TxCnt; i++ {
			tmp := d.TXCache[b.TxArray[i]]
			for j := uint32(0); j < b.MemCnt; j++ {
				tmp.Decision[shard.GlobalGroupMems[b.MemD[j].ID].InShardId] = (b.MemD[j].Decision[index]>>shift)&1 + 1
			}
			tmpRes := b.Result(i)
			if tmpRes == false {
				for j := uint32(0); j < gVar.ShardSize; j++ {
					if tmp.Decision[j] == 1 {
						shard.GlobalGroupMems[b.MemD[j].ID].Rep += gVar.RepTN * int64(tmp.Value)
					} else if tmp.Decision[j] == 2 {
						shard.GlobalGroupMems[b.MemD[j].ID].Rep -= gVar.RepFP * int64(tmp.Value)
					}
				}
			}
			if shift < 7 {
				shift++
			} else {
				index++
				shift = 0
			}
			d.TXCache[b.TxArray[i]] = tmp
		}

	}
	fmt.Println("Leader", d.ID, "get TDS from", b.ID, "with", b.TxCnt, "Txs")
	for i := uint32(0); i < b.TxCnt; i++ {
		tmpHash := b.TxArray[i]
		tmp, ok := d.TXCache[tmpHash]
		//fmt.Println(tmp.InCheck, " ", tmp.InCheckSum)
		if !ok {
			tmp = new(CrossShardDec)
			tmp.NewFromOther(b.ShardIndex, b.Result(i))
			d.TXCache[tmpHash] = tmp
		} else {
			//tmp.Print()
			//fmt.Println("result: ", b.Result(i))
			tmp.UpdateFromOther(b.ShardIndex, b.Result(i))
			//fmt.Println(base58.Encode(tmp.Data.Hash[:]), "result is", tmp.Res)
			if tmp.Res == 1 {
				d.Ready = append(d.Ready, *(tmp.Data))
				for j := uint32(0); j < gVar.ShardSize; j++ {
					if tmp.Decision[j] == 1 {
						shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep -= gVar.RepFN * int64(tmp.Value)
					} else if tmp.Decision[j] == 2 {
						shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep += gVar.RepTP * int64(tmp.Value)
					}
				}
			}
			if tmp.Total == 0 {
				delete(d.TXCache, tmpHash)
			} else {
				d.TXCache[tmpHash] = tmp
			}
		}
	}
	if b.ShardIndex == d.ShardNum {
		b.TxCnt = 0
		b.TxArray = nil
	}
}

//Release delete the first element of the cache
func (d *DbRef) Release() {
	delete(d.TLIndex, d.TLSCache[0][d.ShardNum].HashID)
	//d.TLCache = d.TLCache[1:]
	d.TDSCache = d.TDSCache[1:]
	d.TLSCache = d.TLSCache[1:]
	d.StartIndex++
}
