package rccache

import (
	"github.com/uchihatmtkinu/RC/basic"
)

//GetTxList is to process the TxListX
func (d *DbRef) GetTxList(b *basic.TxListX) *basic.TxList {
	tmp := new(basic.TxList)
	tmp.ID = b.ID
	tmp.HashID = b.HashID
	tmp.TxCnt = b.TxCnt
	tmp.TxArray = make([][32]byte, tmp.TxCnt)
	for i := uint32(0); i < tmp.TxCnt; i++ {
		xxx, ok := d.HashCache[b.TxArray[i]]
		if !ok {
			//ToDo
		} else {
			tmp.TxArray[i] = d.TXCache[xxx[0]]
		}
	}
	tmp.Sig.New(b.Sig)
}
