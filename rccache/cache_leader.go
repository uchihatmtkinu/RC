package rccache

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/uchihatmtkinu/RC/basic"
)

//GenerateTXList is to create TxList given transaction
func (d *dbRef) MakeTXList(a *basic.TxList, b *basic.Transaction) error {
	if b.HashTx() != b.Hash {
		return fmt.Errorf("Hash value invalid")
	}
	if uint32(len(b.In)) != b.TxinCnt || uint32(len(b.Out)) != b.TxoutCnt {
		return fmt.Errorf("Invalid parameter")
	}
	var tmp basic.CrossShardDec
	tmp.New(b)
	if !tmp.InCheck[d.ShardNum] {
		return fmt.Errorf("Not related TX")
	}
	a.AddTx(b)
	d.TX[b.Hash] = &tmp
	for i := uint32(0); i < basic.ShardCnt; i++ {
		if tmp.InCheck[i] {
			d.TL[i].AddTx(b)
		}
	}
	return nil
}

//NewTxList initialize the txList
func (d *dbRef) NewTxList() error {
	tmp := d.TL
	d.TL = make([]basic.TxList, 0, basic.ShardCnt)
	for i := uint32(0); i < basic.ShardCnt; i++ {
		if i != d.ShardNum {
			d.TL[i].ID = d.ID
			if tmp != nil {
				d.TL[i].PrevHash = tmp[i].PrevHash
			} else {
				d.TL[i].PrevHash = sha256.Sum256([]byte(basic.GenesisTL))
			}
		}
	}
	return nil
}

//BuildTxDecSet initialize the txdecset
func (d *dbRef) NewTxDecSet() error {
	d.TDS = make([]basic.TxDecSet, 0, basic.ShardCnt)
	for i := uint32(0); i < basic.ShardCnt; i++ {
		if i != d.ShardNum {
			d.TDS[i].Set(&d.TL[i])
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
	var tmp [32]byte
	var height uint32
	if d.TxB == nil {
		tmp = sha256.Sum256([]byte(basic.GenesisTxBlock))
		height = 1
		d.TxB = new(basic.TxBlock)
	} else {
		tmp = d.TxB.HashID
		height = d.TxB.Height
	}

	d.TxB.MakeTxBlock(d.ID, &d.Ready, tmp, &d.prk, height+1)

	return nil
}

func (d *dbRef) FindNewestTxBlock() error {
	var err error
	d.Mem, err = bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer d.Mem.Close()
	return nil
}
