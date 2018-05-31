package rccache

import "github.com/uchihatmtkinu/RC/basic"

//New initiate the TxDB struct
func (t *CrossShardDec) New(a *basic.Transaction) {
	t.Data = *a
	t.Res = 0
	t.InCheck = make([]bool, basic.ShardCnt)
	for i := uint32(0); i < basic.ShardCnt; i++ {
		t.InCheck[i] = true
	}
	t.InCheckSum = 0
	for i := uint32(0); i < a.TxinCnt; i++ {
		if t.InCheck[a.In[i].ShardIndex()] == true {
			t.InCheck[a.In[i].ShardIndex()] = false
			t.InCheckSum++
		}
	}

}
