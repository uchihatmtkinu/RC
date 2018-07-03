package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//MakeTXList is to create TxList given transaction
func (d *DbRef) MakeTXList(b *basic.Transaction) error {
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
		d.TLIndex[d.TLS[d.ShardNum].Hash()] = uint32(d.LastIndex)
	}

	d.TLS = new([gVar.ShardCnt]basic.TxList)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.TLS[i].ID = d.ID
	}
	//d.TL = new(basic.TxList)
	//d.TL.Set(d.ID)
	return nil
}

//GenerateTxBlock makes the TxBlock
func (d *DbRef) GenerateTxBlock() error {
	height := d.TxB.Height
	d.TxB.MakeTxBlock(d.ID, &d.Ready, d.DB.LastTB, &d.prk, height+1, 0, nil, 0)
	for i := 0; i < len(d.Ready); i++ {
		d.ClearCache(d.Ready[i].Hash)
	}
	d.Ready = nil
	d.DB.AddBlock(d.TxB)
	d.DB.UpdateUTXO(d.TxB, d.ShardNum)
	return nil
}

//GenerateFinalBlock generate final block
func (d *DbRef) GenerateFinalBlock() error {
	tmp := d.DB.MakeFinalTx()
	height := d.TxB.Height
	d.TxB.MakeTxBlock(d.ID, tmp, d.DB.LastFB[d.ShardNum], &d.prk, height+1, 1, &d.FB[d.ShardNum].HashID, d.ShardNum)
	return nil
}

//GenerateStartBlock generate Start block
func (d *DbRef) GenerateStartBlock() error {
	height := d.FB[d.ShardNum].Height
	tmp := make([][32]byte, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		tmp[i] = d.FB[i].HashID
	}
	d.TxB.MakeStartBlock(d.ID, &tmp, d.DB.LastFB[d.ShardNum], &d.prk, height+1)
	return nil
}

//UpdateTXCache is to pick the transactions into ready slice given txdecision
func (d *DbRef) UpdateTXCache(a *basic.TxDecision) error {
	if a.Single == 1 {
		return fmt.Errorf("TxDecision parameter error")
	}
	err := d.PreTxDecision(a, a.HashID)
	if err != nil {
		return err
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
	for i := uint32(0); i < tmpTL.TxCnt; i++ {
		tmpTx := d.TXCache[tmpTL.TxArray[i]]
		for j := 0; j < len(tmpTx.ShardRelated); j++ {
			tmpTD[j].Add((a.Decision[x] >> y) & 1)
		}
		if y < 7 {
			y++
		} else {
			x++
			y = 0
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
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
	}
	for i := uint32(0); i < b.TxCnt; i++ {
		tmpHash := b.TxArray[i]
		tmp, ok := d.TXCache[tmpHash]
		//fmt.Println(tmp.InCheck, " ", tmp.InCheckSum)
		if !ok {
			tmp = new(CrossShardDec)
			tmp.NewFromOther(b.ShardIndex, b.Result(i))
			d.TXCache[tmpHash] = tmp
		} else {
			tmp.UpdateFromOther(b.ShardIndex, b.Result(i))
			if tmp.Res == 1 {
				d.Ready = append(d.Ready, *(tmp.Data))
			}
			if tmp.Total == 0 {
				delete(d.TXCache, tmpHash)
			} else {
				d.TXCache[tmpHash] = tmp
			}
		}
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
