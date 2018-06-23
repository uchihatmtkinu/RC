package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/gVar"

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
	if !b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk) {
		return fmt.Errorf("Signature not match")
	}
	return nil
}

//PreTxDecision is verify the txdecision
func (d *DbRef) PreTxDecision(b *basic.TxDecision, hash [32]byte) error {
	if int(b.TxCnt) > len(b.Decision)*8 {
		return fmt.Errorf("Decision lengh not enough")
	}
	if b.Single == 0 {
		if b.Target != d.ShardNum {
			return fmt.Errorf("TxDecision should be the intra-one")
		}
		if shard.GlobalGroupMems[b.ID].Shard != int(d.ShardNum) {
			return fmt.Errorf("Not the same shard")
		}
		if len(b.Sig) != int(gVar.ShardCnt) {
			return fmt.Errorf("Signature not enough")
		}
		if !b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk, d.ShardNum) {
			return fmt.Errorf("Signature not match")
		}
	} else {
		b.HashID = hash
		if b.Target != d.ShardNum {
			return fmt.Errorf("Not the target shard")
		}
		if len(b.Sig) != 1 {
			return fmt.Errorf("Signature not enough")
		}
		if !b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk, 0) {
			return fmt.Errorf("Signature not match")
		}
	}
	return nil
}

//PreTxDecSet is to process the TxListX
func (d *DbRef) PreTxDecSet(b *basic.TxDecSet) error {
	if shard.GlobalGroupMems[b.ID].Role != 0 {
		return fmt.Errorf("Not a Leader")
	}
	if int(b.TxCnt) != len(b.TxArrayX) || int(b.MemCnt) != len(b.MemD) {
		return fmt.Errorf("TxDecSet parameter not match")
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
	for i := uint32(0); i < b.MemCnt; i++ {
		err := d.PreTxDecision(&b.MemD[i], b.HashID)
		if err != nil {
			return err
		}
	}
	if !b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk) {
		return fmt.Errorf("Signature not match")
	}
	return nil
}

//PreTxBlock is to process the TxListX
func (d *DbRef) PreTxBlock(b *basic.TxBlock) error {
	if shard.GlobalGroupMems[b.ID].Role != 0 {
		return fmt.Errorf("Not a Leader")
	}
	if int(b.TxCnt) != len(b.TxArrayX) {
		return fmt.Errorf("TxBlock parameter not match")
	}
	b.TxArray = make([]basic.Transaction, b.TxCnt)
	for i := uint32(0); i < b.TxCnt; i++ {
		xxx, ok := d.HashCache[b.TxArrayX[i]]
		if !ok {
			//ToDo
		} else {
			b.TxArray[i] = *d.TXCache[xxx[0]].Data
		}
	}
	tmp, err := b.Verify(&shard.GlobalGroupMems[b.ID].RealAccount.Puk)
	if !tmp {
		return err
	}
	return nil
}
