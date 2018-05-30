package rccache

import (
	"fmt"

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
	a.AddTx(b)
	return nil
}

//BuildTxDecSet initialize the txdecset given txlist
func (d *dbRef) BuildTxDecSet(a *[]basic.TxDecSet, b *basic.TxList) error {
	return nil
}

//UpdateTxDecSet given TxDecision
func (d *dbRef) UpdateTxDecSet(a *basic.TxDecSet, b *basic.TxDecision) error {
	return nil
}
