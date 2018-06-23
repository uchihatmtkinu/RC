package basic

import "fmt"

//Transform change the txlist to txlistX
func (a *TxList) Transform() *TxListX {
	tmp := new(TxListX)
	tmp.ID = a.ID
	tmp.HashID = a.HashID
	tmp.Sig.New(&a.Sig)
	tmp.TxCnt = a.TxCnt
	tmp.TxArray = make([][sHash]byte, a.TxCnt)
	for i := uint32(0); i < tmp.TxCnt; i++ {
		tmp.TxArray[i] = HashCut(a.TxArray[i])
	}
	return tmp
}

//Encode returns the byte of a TxList
func (a *TxListX) Encode(tmp *[]byte) {
	Encode(tmp, a.ID)
	Encode(tmp, &a.HashID)
	Encode(tmp, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		Encode(tmp, a.TxArray[i])
	}
	Encode(tmp, &a.Sig)
}

//Decode decodes the TxList with []byte
func (a *TxListX) Decode(buf *[]byte) error {
	err := Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("TxList ID decode failed: %s", err)
	}
	err = Decode(buf, &a.HashID)
	if err != nil {
		return fmt.Errorf("TxList HashID decode failed: %s", err)
	}
	err = Decode(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxList TxCnt decode failed: %s", err)
	}
	a.TxArray = make([][sHash]byte, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		err = Decode(buf, &a.TxArray[i])
		if err != nil {
			return fmt.Errorf("TxList Tx decode failed-%d: %s", i, err)
		}
	}
	err = Decode(buf, &a.Sig)
	if err != nil {
		return fmt.Errorf("TxList signature decode failed: %s", err)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxList decode failed: With extra bits")
	}
	return nil
}
