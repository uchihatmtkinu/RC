package basic

//New initiate the TxDB struct
func (t *TxDB) New(a *Transaction) {
	t.Data = *a
	t.Res = 0
	t.Used = make([]uint32, a.TxoutCnt)
	t.InCheck = make([]bool, a.TxinCnt)
	for i := uint32(0); i < a.TxinCnt; i++ {
		t.InCheck[i] = false
	}
	t.InCheckSum = 0
}
