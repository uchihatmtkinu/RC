package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
)

//MakeTXList is to create TxList given transaction
func (d *DbRef) MakeTXList(b *basic.Transaction) error {
	tmpHash := HashCut(b.Hash)
	tmp, ok := d.TXCache[tmpHash]
	if !ok {
		tmp.New(b)
	} else {
		tmp = new(CrossShardDec)
		tmp.Update(b)
	}
	if tmp.InCheck[d.ShardNum] == 0 {
		if ok {
			delete(d.TXCache, tmpHash)
		}
		return fmt.Errorf("Not related TX")
	}
	d.TL.AddTx(b)
	d.TXCache[tmpHash] = tmp
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if tmp.InCheck[i] != 0 {
			d.TLS[i].AddTx(b)
		}
	}
	return nil
}

//SignTXL is to sign all txlist
func (d *DbRef) SignTXL() {
	d.TL.Sign(&d.prk)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if i != d.ShardNum {
			d.TLS[i].Sign(&d.prk)
		}
	}
}

//BuildTDS is to sign all txDecSet
//Must after SignTXL
func (d *DbRef) BuildTDS() {
	d.TDS = new([gVar.ShardCnt]basic.TxDecSet)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.TDS[i].Set(&d.TLS[i], d.ShardNum)
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.TDS[i].Sign(&d.prk)
	}
}

//NewTxList initialize the txList
//Must after BuildTDS
func (d *DbRef) NewTxList() error {
	if d.TL != nil {
		d.TLCache = append(d.TLCache, *d.TL)
		d.TLSCache = append(d.TLSCache, *d.TLS)
		d.TDSCache = append(d.TDSCache, *d.TDS)
		d.lastIndex++
		d.TLIndex[d.TL.Hash()] = uint32(d.lastIndex)
	}

	d.TLS = new([gVar.ShardCnt]basic.TxList)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.TLS[i].ID = d.ID
	}
	d.TL = new(basic.TxList)
	d.TL.Set(d.ID)
	return nil
}

//GenerateTxBlock makes the TxBlock
func (d *DbRef) GenerateTxBlock() error {
	height := d.TxB.Height
	d.TxB.MakeTxBlock(d.ID, &d.Ready, d.db.lastTB, &d.prk, height+1, 0)
	for i := 0; i < len(d.Ready); i++ {
		delete(d.TXCache, HashCut(d.Ready[i].Hash))
	}
	d.Ready = nil
	d.db.AddBlock(d.TxB)
	d.db.UpdateUTXO(d.TxB, d.ShardNum)
	return nil
}

//GenerateFinalBlock generate final block
func (d *DbRef) GenerateFinalBlock() error {
	tmp := d.db.MakeFinalTx()
	height := d.TxB.Height
	d.TxB.MakeTxBlock(d.ID, tmp, d.db.lastTB, &d.prk, height+1, 1)
	return nil
}

//UpdateTXCache is to pick the transactions into ready slice given txdecision
func (d *DbRef) UpdateTXCache(a *basic.TxDecision) error {
	if a.Target != d.ShardNum {
		return fmt.Errorf("TxDecision should be the intra-one")
	}
	if a.Single != 1 || uint32(len(a.Sig)) != gVar.ShardCnt {
		return fmt.Errorf("TxDecision parameter error")
	}
	tmp, ok := d.TLIndex[a.HashID]
	if !ok {
		return fmt.Errorf("TxDecision Hash error, wrong or time out")
	}
	tmpIndex := tmp - uint32(d.startIndex)
	tmpTL := d.TLCache[tmpIndex]

	var x, y uint32 = 0, 0
	tmpTD := make([]basic.TxDecision, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {

		tmpTD[i].Set(a.ID, i, 1)
		tmpTD[i].HashID = d.TLSCache[tmpIndex][i].HashID
		tmpTD[i].Single = 1
		tmpTD[i].Sig = nil
		tmpTD[i].Sig = append(tmpTD[i].Sig, a.Sig[i])

	}
	for i := uint32(0); i < tmpTL.TxCnt; i++ {
		tmpTx := d.TXCache[HashCut(tmpTL.TxArray[i])]
		for j := 0; j < len(tmpTx.ShardRelated); j++ {
			tmpTD[j].Add((a.Decision[x] >> y) & 1)
		}
		if y < 7 {
			y++
		} else {
			x++
			if x >= uint32(len(a.Decision)) {
				break
			}
			y = 0
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.TDS[i].Add(&tmpTD[i])
	}
	return nil
}

//ProcessTDS deal with the TDS
func (d *DbRef) ProcessTDS(b *basic.TxDecSet) {
	if b.ShardIndex != d.ShardNum {

	}
	for i := uint32(0); i < b.TxCnt; i++ {
		tmpHash := HashCut(b.TxArray[i])
		tmp, ok := d.TXCache[tmpHash]
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
	delete(d.TLIndex, d.TLCache[0].HashID)
	d.TLCache = d.TLCache[1:]
	d.TDSCache = d.TDSCache[1:]
	d.TLSCache = d.TLSCache[1:]
	d.startIndex++
}
