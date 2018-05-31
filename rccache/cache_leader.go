package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/basic"
)

//GenerateTXList is to create TxList given transaction
func (d *dbRef) MakeTXList(b *basic.Transaction) error {
	if b.HashTx() != b.Hash {
		return fmt.Errorf("Hash value invalid")
	}
	if uint32(len(b.In)) != b.TxinCnt || uint32(len(b.Out)) != b.TxoutCnt {
		return fmt.Errorf("Invalid parameter")
	}
	tmp, ok := d.TXCache[b.Hash]
	if !ok {
		tmp.New(b)
	} else {
		tmp.Update(b)
	}
	if tmp.InCheck[d.ShardNum] == 0 {
		if ok {
			delete(d.TXCache, b.Hash)
		}
		return fmt.Errorf("Not related TX")
	}
	d.TL.AddTx(b)
	d.TXCache[b.Hash] = tmp
	for i := uint32(0); i < basic.ShardCnt; i++ {
		if tmp.InCheck[i] > 0 {
			d.TLS[i].AddTx(b)
		}
	}
	return nil
}

//NewTxList initialize the txList
func (d *dbRef) NewTxList() error {
	if d.TL != nil {
		d.TLCache = append(d.TLCache, *d.TL)
		d.TLSCache = append(d.TLSCache, *d.TLS)
		d.TDSCache = append(d.TDSCache, *d.TDS)
		d.lastIndex++
		d.TLIndex[d.TL.Hash()] = uint32(d.lastIndex)
	}

	d.TLS = new([basic.ShardCnt]basic.TxList)
	for i := uint32(0); i < basic.ShardCnt; i++ {
		if i != d.ShardNum {
			d.TLS[i].ID = d.ID
			d.TLS[i].PrevHash = d.db.lastTLS[i]
		}
	}
	d.TL.Set(d.ID, d.db.lastTL)

	d.TDS = new([basic.ShardCnt]basic.TxDecSet)
	for i := uint32(0); i < basic.ShardCnt; i++ {
		if i != d.ShardNum {
			d.TDS[i].Set(&d.TLS[i], d.ShardNum)
		} else {
			d.TDS[i].Set(d.TL, d.ShardNum)
		}
	}
	return nil
}

//GenerateTxBlock makes the TxBlock
func (d *dbRef) GenerateTxBlock() error {
	height := d.TxB.Height
	d.TxB.MakeTxBlock(d.ID, &d.Ready, d.db.lastTB, &d.prk, height+1, 0)
	for i := 0; i < len(d.Ready); i++ {
		delete(d.TXCache, d.Ready[i].Hash)
	}
	d.Ready = nil
	d.db.AddBlock(d.TxB)
	d.db.UpdateUTXO(d.TxB)
	return nil
}

func (d *dbRef) GenerateFinalBlock() error {
	tmp := d.db.MakeFinalTx()
	height := d.TxB.Height
	d.TxB.MakeTxBlock(d.ID, tmp, d.db.lastTB, &d.prk, height+1, 1)
	return nil
}

//UpdateTXCache is to pick the transactions into ready slice given txdecision
func (d *dbRef) UpdateTXCache(a *basic.TxDecision) error {
	if a.Target != d.ShardNum {
		return fmt.Errorf("TxDecision should be the intra-one")
	}
	if a.Single != 1 || uint32(len(a.Sig)) != basic.ShardCnt {
		return fmt.Errorf("TxDecision parameter error")
	}
	tmp, ok := d.TLIndex[a.HashID]
	if !ok {
		return fmt.Errorf("TxDecision Hash error, wrong or time out")
	}
	tmpIndex := tmp - uint32(d.startIndex)
	tmpTL := d.TLCache[tmpIndex]

	var x, y uint32 = 0, 0
	tmpTD := make([]basic.TxDecision, basic.ShardCnt)
	for i := uint32(0); i < basic.ShardCnt; i++ {

		tmpTD[i].Set(a.ID, &d.TLSCache[tmpIndex][i], i)
		tmpTD[i].Single = 1
		tmpTD[i].Sig = nil
		tmpTD[i].Sig = append(tmpTD[i].Sig, a.Sig[i])

	}
	for i := uint32(0); i < tmpTL.TxCnt; i++ {
		tmpTx := d.TXCache[tmpTL.TxArray[i].Hash]
		if (a.Decision[x]>>y)&1 == 1 {
			tmpTx.Yes++

		} else {
			tmpTx.No++
		}

		if tmpTx.Yes > (basic.ShardSize-1)/2 {
			tmpTx.UpdateFromOther(d.ShardNum, true)
		} else if tmpTx.No > (basic.ShardSize-1)/2 {
			tmpTx.UpdateFromOther(d.ShardNum, false)
		}
		d.TXCache[tmpTL.TxArray[i].Hash] = tmpTx
		for j := 0; j < len(tmpTx.ShardRelated); j++ {

			tmpTD[j].Add((a.Decision[x] >> y) & 1)

		}
		if y < 7 {
			y++
		} else {
			x++
			y = 0
		}
		if x >= uint32(len(a.Decision)) {
			break
		}
	}
	for i := uint32(0); i < basic.ShardCnt; i++ {

		d.TDS[i].Add(&tmpTD[i])

	}
	return nil
}

//ProcessTDS deal with the TDS
func (d *dbRef) ProcessTDS(b *basic.TxDecSet) {
	if b.ShardIndex != d.ShardNum {
		if d.TDSS.TxCnt >= basic.MaxDecSSSize {
			d.TDSSSent = d.TDSS
			tmp := []basic.TxDecSet{*b}
			d.TDSS.Build(&tmp)
		} else {
			d.TDSS.AddTxDec(b)
		}
	}
	for i := uint32(0); i < b.TxCnt; i++ {
		tmp, ok := d.TXCache[b.TxArray[i]]
		if !ok {
			tmp.NewFromOther(b.ShardIndex, b.Result(i))
			d.TXCache[b.TxArray[i]] = tmp
		} else {
			tmp.UpdateFromOther(b.ShardIndex, b.Result(i))
			if tmp.Res == 1 {
				d.Ready = append(d.Ready, *(tmp.Data))
			}
			if tmp.Total == 0 {
				delete(d.TXCache, b.TxArray[i])
			} else {
				d.TXCache[b.TxArray[i]] = tmp
			}
		}
	}
}

//Release delete the first element of the cache
func (d *dbRef) Release() {
	delete(d.TLIndex, d.TLCache[0].HashID)
	d.TLCache = d.TLCache[1:]
	d.TDSCache = d.TDSCache[1:]
	d.TLSCache = d.TLSCache[1:]
	d.startIndex++
}
