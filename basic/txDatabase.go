package basic

//New initiate the TxDB struct
func (t *TxDB) New(a *Transaction) {
	t.ID = a.Hash
	t.Res = 0
	t.Used = make([]uint32, a.TxoutCnt)
	t.InCheck = make([]bool, ShardCnt)
	for i := uint32(0); i < ShardCnt; i++ {
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
