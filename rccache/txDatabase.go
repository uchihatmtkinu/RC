package rccache

import "github.com/uchihatmtkinu/RC/basic"

//New initiate the TxDB struct
func (t *CrossShardDec) New(a *basic.Transaction) {
	newT := *a
	t.Data = &newT
	t.Res = 0
	t.InCheck = make([]int, basic.ShardCnt)
	tmp := make([]bool, basic.ShardCnt)
	t.InCheckSum = 0

	for i := uint32(0); i < a.TxinCnt; i++ {
		xx := a.In[i].ShardIndex()
		tmp[xx] = true
		t.InCheck[xx] = 3
	}
	for i := uint32(0); i < a.TxoutCnt; i++ {
		tmp[a.Out[i].ShardIndex()] = true
	}

	t.ShardRelated = make([]uint32, 0, basic.ShardCnt)
	for i := uint32(0); i < basic.ShardCnt; i++ {
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
	tmp := make([]bool, basic.ShardCnt)
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
	}

	t.ShardRelated = make([]uint32, 0, basic.ShardCnt)
	for i := uint32(0); i < basic.ShardCnt; i++ {
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
	t.InCheck = make([]int, basic.ShardCnt)

	t.Total = int(basic.ShardCnt - 1)
	if res {
		t.Res = 0
		t.InCheck[index] = 1
		t.InCheckSum = int(basic.ShardCnt - 1)
	} else {
		t.Res = -1
		t.InCheck[index] = 2
		t.InCheckSum = int(basic.ShardCnt)
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
