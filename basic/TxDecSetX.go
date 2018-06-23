package basic

import "fmt"

//Transform change the txlist to txlistX
func (a *TxDecSet) Transform() *TxDecSetX {
	tmp := new(TxDecSetX)
	tmp.ID = a.ID
	tmp.HashID = a.HashID
	tmp.MemCnt = a.MemCnt
	tmp.ShardIndex = a.ShardIndex
	tmp.MemD = make([]TxDecision, tmp.MemCnt)
	copy(tmp.MemD, a.MemD)
	tmp.TxCnt = a.TxCnt
	tmp.TxArray = make([][sHash]byte, a.TxCnt)
	for i := uint32(0); i < tmp.TxCnt; i++ {
		tmp.TxArray[i] = HashCut(a.TxArray[i])
	}
	tmp.Sig.New(&a.Sig)
	return tmp
}

//Encode encode the TxDecSet into []byte
func (a *TxDecSetX) Encode(tmp *[]byte) {
	Encode(tmp, a.ID)
	Encode(tmp, &a.HashID)
	Encode(tmp, a.MemCnt)
	Encode(tmp, a.TxCnt)
	Encode(tmp, a.ShardIndex)
	for i := uint32(0); i < a.MemCnt; i++ {
		a.MemD[i].Encode(tmp)
	}
	for i := uint32(0); i < a.TxCnt; i++ {
		Encode(tmp, &a.TxArray[i])
	}
	Encode(tmp, &a.Sig)
}

//Decode decode the []byte into TxDecSet
func (a *TxDecSetX) Decode(buf *[]byte) error {
	err := Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("TxDecSet ID decode failed: %s", err)
	}
	err = Decode(buf, &a.HashID)
	if err != nil {
		return fmt.Errorf("TxDecSet HashID decode failed: %s", err)
	}
	err = Decode(buf, &a.MemCnt)
	if err != nil {
		return fmt.Errorf("TxDecSet MemCnt decode failed: %s", err)
	}
	err = Decode(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxDecSet TxCnt decode failed: %s", err)
	}
	err = Decode(buf, &a.ShardIndex)
	if err != nil {
		return fmt.Errorf("TxDecSet ShardIndex decode failed: %s", err)
	}
	a.MemD = make([]TxDecision, a.MemCnt)
	for i := uint32(0); i < a.MemCnt; i++ {
		err = a.MemD[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("TxDecSet MemDecision decode failed-%d: %s", i, err)
		}
	}
	a.TxArray = make([][sHash]byte, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		err = Decode(buf, &a.TxArray[i])
		if err != nil {
			return fmt.Errorf("TxDecSet TxArray decode failed-%d: %s", i, err)
		}
	}
	err = Decode(buf, &a.Sig)
	if err != nil {
		return fmt.Errorf("TxDecSet Signature decode failed: %s", err)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxDecSet decode failed: With extra bits")
	}
	return nil
}
