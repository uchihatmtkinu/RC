package rccache

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/basic"
)

//tmpTx stores the tempory utxo information
type tmpTx struct {
	index uint32
	under uint32
	value uint32
}

//dbRef is the structure stores the cache of a miner for the database
type dbRef struct {
	TXIndex map[[32]byte]int
	TX      *[]basic.TxDB
	tmp     *[]tmpTx
	TXLen   uint32
}

func (d *dbRef) findTX(a *basic.Transaction, index int) int {

	x, ok := d.TXIndex[a.In[index].PrevTx]
	if !ok {
		return -1
	}
	if (*d.TX)[x].Data.Kind == 1 {
		return 0
	}
	if (*d.TX)[x].Data.TxoutCnt < a.In[index].Index {
		return -1
	}
	return x

}

func (d *dbRef) AddTX(a *basic.Transaction) {
	var xxx basic.TxDB
	xxx.Data = *a
	xxx.Used = make([]uint32, a.TxoutCnt)
	xxx.Res = -1
	(*d.TX) = append(*d.TX, xxx)
	d.TXIndex[xxx.Data.Hash] = len(*d.TX) - 1
	d.TXLen++
}

func (d *dbRef) Verify(a *basic.Transaction) (bool, error) {
	if a.TxinCnt != uint32(len(a.In)) || a.TxoutCnt != uint32(len(a.Out)) {
		return false, fmt.Errorf("Invalid input,output parameter")
	}
	tmp := a.HashTx()
	if tmp != a.Hash {
		return false, fmt.Errorf("Hashvalue not match")
	}
	d.AddTX(a)
	if a.Kind == 0 {
		var value uint32
		var tmpInt uint32
		var tmpOut basic.OutType
		tmpArr := make([]uint32, a.TxinCnt)
		check := true
		for i := uint32(0); i < a.TxinCnt; i++ {
			x := d.findTX(a, int(i))
			if x < 0 {
				return false, fmt.Errorf("rccache.Verify Invalid UTXO of %d", &i)
			}
			tmpArr[i] = uint32(x)
			if (*d.TX)[x].Data.Kind == 1 {
				tmpOut = (*d.TX)[x].Data.Out[0]
				if !a.VerifyTx(i, &tmpOut) {
					return false, fmt.Errorf("rccache.Verify Invalid UTXO of %d address", &i)
				}
				if (*d.TX)[x].Used[0]+a.In[i].Index > (*d.TX)[x].Data.Out[0].Value {
					check = false
				}
				tmpInt = a.In[i].Index
			} else {
				tmpOut = (*d.TX)[x].Data.Out[a.In[i].Index]
				if !a.VerifyTx(i, &tmpOut) {
					return false, fmt.Errorf("rccache.Verify Invalid UTXO of %d address", &i)
				}
				if (*d.TX)[x].Used[a.In[i].Index] != 0 {
					check = false
				}
				tmpInt = (*d.TX)[x].Data.Out[a.In[i].Index].Value
			}

			value += tmpInt
		}
		for i := uint32(0); i < a.TxoutCnt; i++ {
			value -= a.Out[i].Value
		}
		if value < 0 {
			return false, fmt.Errorf("rccache.Verify Invalid outcome value")
		}

		if check {
			(*d.TX)[len(*d.TX)-1].Res = 1
			for i := uint32(0); i < a.TxinCnt; i++ {
				if (*d.TX)[tmpArr[i]].Data.Kind == 0 {
					(*d.TX)[tmpArr[i]].Used[a.In[i].Index] = d.TXLen

				} else {
					(*d.TX)[tmpArr[i]].Used[0] += a.In[i].Index
				}
				var tmpTX0 tmpTx
				tmpTX0.index = tmpArr[i]
				tmpTX0.under = a.In[i].Index
				tmpTX0.value = d.TXLen
				*d.tmp = append(*d.tmp, tmpTX0)
			}
			return true, nil
		}
		(*d.TX)[len(*d.TX)-1].Res = -2
		return false, fmt.Errorf("rccache.Verify UTXO used")

	} else if a.Kind == 1 {
		if a.TxoutCnt != 1 {
			return false, fmt.Errorf("rccache.Verify the out address should be 1")
		}

	}
	return false, fmt.Errorf("rccache.Verify Invalid transaction type")
}
