package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/shard"
)

//PreTxList is to process the TxListX
func (d *DbRef) PreTxList(b *basic.TxList) error {
	if d.Leader != b.ID {
		return fmt.Errorf("Txlist from a miner")
	}
	if int(b.TxCnt) != len(b.TxArrayX) {
		return fmt.Errorf("Number of Tx wrong")
	}
	b.TxArray = make([][32]byte, b.TxCnt)
	for i := uint32(0); i < b.TxCnt; i++ {
		xxx, ok := d.HashCache[b.TxArrayX[i]]
		if !ok {
			//ToDo
		} else {
			b.TxArray[i] = d.TXCache[xxx[0]].Data.Hash
		}
	}
	if b.Hash() != b.HashID {
		return fmt.Errorf("Hash not match")
	}
	if b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk) {
		return fmt.Errorf("Signature not match")
	}
	return nil
}

//PreTxDecSet is to process the TxListX
func (d *DbRef) PreTxDecSet(b *basic.TxDecSet) error {
	b.TxArray = make([][32]byte, b.TxCnt)
	for i := uint32(0); i < b.TxCnt; i++ {
		xxx, ok := d.HashCache[b.TxArrayX[i]]
		if !ok {
			//ToDo
		} else {
			b.TxArray[i] = d.TXCache[xxx[0]].Data.Hash
		}
	}
	return nil
}

//PreTxBlock is to process the TxListX
func (d *DbRef) PreTxBlock(b *basic.TxBlock) error {
	b.TxArray = make([]basic.Transaction, b.TxCnt)
	for i := uint32(0); i < b.TxCnt; i++ {
		xxx, ok := d.HashCache[b.TxArrayX[i]]
		if !ok {
			//ToDo
		} else {
			b.TxArray[i] = *d.TXCache[xxx[0]].Data
		}
	}
	return nil
}
