package rccache

import (
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
)

//New initiate the TxDB struct
func (t *CrossShardDec) New(a *basic.Transaction) {
	newT := *a
	t.Data = &newT
	t.Res = 0
	t.InCheck = make([]int, gVar.ShardCnt)
	tmp := make([]bool, gVar.ShardCnt)
	t.InCheckSum = 0
	for i := uint32(0); i < a.TxinCnt; i++ {
		xx := a.In[i].ShardIndex()
		tmp[xx] = true
		t.InCheck[xx] = 3
	}
	for i := uint32(0); i < a.TxoutCnt; i++ {
		tmp[a.Out[i].ShardIndex()] = true
		t.Value += a.Out[i].Value
	}

	t.ShardRelated = make([]uint32, 0, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if tmp[i] {
			if t.InCheck[i] == 0 {
				t.InCheck[i] = -1
			}
			t.ShardRelated = append(t.ShardRelated, i)
		}
		if t.InCheck[i] == 3 {
			t.InCheckSum++
		}
	}
	t.Total = t.InCheckSum
}

//Update from the transaction
func (t *CrossShardDec) Update(a *basic.Transaction) {
	newT := *a
	t.Data = &newT
	tmp := make([]bool, gVar.ShardCnt)
	t.InCheckSum = 0

	for i := uint32(0); i < a.TxinCnt; i++ {
		xx := a.In[i].ShardIndex()
		if t.InCheck[xx] == 0 {
			t.InCheck[xx] = 3
			tmp[xx] = true
		}
	}
	for i := uint32(0); i < a.TxoutCnt; i++ {
		tmp[a.Out[i].ShardIndex()] = true
		t.Value += a.Out[i].Value
	}

	t.ShardRelated = make([]uint32, 0, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		if tmp[i] {
			if t.InCheck[i] == 0 {
				t.InCheck[i] = -1
			}
			t.ShardRelated = append(t.ShardRelated, i)
		}
		if t.InCheck[i] > 1 {
			t.InCheckSum++
		}
	}
	t.Total = t.InCheckSum
}

//NewFromOther initiate the TxDB struct by cross-shard data
func (t *CrossShardDec) NewFromOther(index uint32, res bool) {
	t.Data = nil
	t.Res = 0
	t.InCheck = make([]int, gVar.ShardCnt)

	t.Total = int(gVar.ShardCnt - 1)
	if res {
		t.Res = 0
		t.InCheck[index] = 1
		t.InCheckSum = int(gVar.ShardCnt - 1)
	} else {
		t.Res = -1
		t.InCheck[index] = 2
		t.InCheckSum = int(gVar.ShardCnt)
	}
}

//UpdateFromOther initiate the TxDB struct by cross-shard data
func (t *CrossShardDec) UpdateFromOther(index uint32, res bool) {

	if res {
		if t.InCheck[index] == 3 {
			t.InCheck[index] = 1
			t.InCheckSum--
			if t.InCheckSum == 0 {
				t.Res = 1
			}
			t.Total--
		}
	} else {
		if t.InCheck[index] == 3 {
			t.InCheck[index] = 2
			t.Total--
		}
		t.Res = -1
	}

}
