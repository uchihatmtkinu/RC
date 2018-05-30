package basic

import "fmt"

//New initiate the TxDB struct
func (t *CrossShardDec) New(a *Transaction) {
	t.Data = *a
	t.Res = 0
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

//Encode encodes TxDB into []byte
func (t *TxDB) Encode(tmp *[]byte) {
	t.Data.Encode(tmp)
	EncodeByte(tmp, &t.Used)
}

//Decode decodes the []byte into TxDB
func (t *TxDB) Decode(buf *[]byte) error {
	err := t.Data.Decode(buf)
	if err != nil {
		return fmt.Errorf("TxDB data decoding failed")
	}
	err = DecodeByte(buf, &t.Used)
	if err != nil {
		return fmt.Errorf("TxDB byte array decoding failed")
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxDB decoding remains bit")
	}
	return nil
}

//Lock is assign the value to 2
func (t *TxDB) Lock(index uint32) {
	t.Used[index] = 2
}
