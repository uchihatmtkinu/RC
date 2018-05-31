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
	var tmp CrossShardDec
	tmp.New(b)
	if !tmp.InCheck[d.ShardNum] {
		return fmt.Errorf("Not related TX")
	}
	d.TL.AddTx(b)
	d.TX[b.Hash] = &tmp
	for i := uint32(0); i < basic.ShardCnt; i++ {
		if tmp.InCheck[i] {
			d.TLS[i].AddTx(b)
		}
	}
	return nil
}

//NewTxList initialize the txList
func (d *dbRef) NewTxList() error {
	d.TLS = make([]basic.TxList, 0, basic.ShardCnt)
	for i := uint32(0); i < basic.ShardCnt; i++ {
		if i != d.ShardNum {
			d.TLS[i].ID = d.ID
			d.TLS[i].PrevHash = d.db.lastTLS[i]
		}
	}
	d.TL.Set(d.ID, d.db.lastTL)
	return nil
}

//BuildTxDecSet initialize the txdecset
func (d *dbRef) NewTxDecSet() error {
	d.TDS = make([]basic.TxDecSet, basic.ShardCnt)
	for i := uint32(0); i < basic.ShardCnt; i++ {
		if i != d.ShardNum {
			d.TDS[i].Set(&d.TLS[i])
		}
	}
	return nil
}

func (d *dbRef) AddTxDec(b *basic.TxDecision, index uint32) error {
	d.TDS[index].Add(b)
	return nil
}

//GenerateTxBlock makes the TxBlock
func (d *dbRef) GenerateTxBlock() error {
	height := d.TxB.Height
	d.TxB.MakeTxBlock(d.ID, &d.Ready, d.db.lastTB, &d.prk, height+1)

	return nil
}
