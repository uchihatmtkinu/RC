package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
)

//Sign signs the TxDecSet
func (a *TxDecSet) Sign(prk *ecdsa.PrivateKey) {
	tmp := make([]byte, 0, 64+len(a.MemD[0].Decision)*int(a.MemCnt))
	tmp = append(a.ID[:], a.HashID[:]...)
	for i := uint32(0); i < a.MemCnt; i++ {
		tmp = append(tmp, a.MemD[i].Decision...)
	}
	tmpHash := sha256.Sum256(tmp)
	a.Sig.Sign(tmpHash[:], prk)
}

//Verify verifies the TxDecSet
func (a *TxDecSet) Verify(puk *ecdsa.PublicKey) bool {
	tmp := make([]byte, 0, 64+len(a.MemD[0].Decision)*int(a.MemCnt))
	tmp = append(a.ID[:], a.HashID[:]...)
	for i := uint32(0); i < a.MemCnt; i++ {
		tmp = append(tmp, a.MemD[i].Decision...)
	}
	tmpHash := sha256.Sum256(tmp)
	return a.Sig.Verify(tmpHash[:], puk)
}

//Set init an instance of TxDecSet given those parameters
func (a *TxDecSet) Set(b *TxList) {
	a.ID = b.ID
	a.HashID = b.HashID
	a.PrevHash = b.PrevHash
	a.MemCnt = 0
	a.TxCnt = b.TxCnt
	for i := uint32(0); i < a.TxCnt; i++ {
		a.TxArray = append(a.TxArray, b.TxArray[i].Hash)
	}
}

//Add adds a TxDecision
func (a *TxDecSet) Add(b *TxDecision) {
	a.MemCnt++
	a.MemD = append(a.MemD, *b)
}

//Encode encode the TxDecSet into []byte
func (a *TxDecSet) Encode(tmp *[]byte) {
	EncodeByteL(tmp, a.ID[:], 32)
	EncodeByteL(tmp, a.ID[:], 32)
	EncodeByteL(tmp, a.ID[:], 32)
	EncodeInt(tmp, a.MemCnt)
	EncodeInt(tmp, a.TxCnt)
	for i := uint32(0); i < a.MemCnt; i++ {
		a.MemD[i].Encode(tmp)
	}
	for i := uint32(0); i < a.TxCnt; i++ {
		*tmp = append(*tmp, a.TxArray[i][:]...)
	}
	a.Sig.SignToData(tmp)
}

//Decode decode the []byte into TxDecSet
func (a *TxDecSet) Decode(buf *[]byte) error {
	var tmp []byte
	err := DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxDecSet ID decode failed %s", err)
	}
	copy(a.ID[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxDecSet HashID decode failed %s", err)
	}
	copy(a.HashID[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxDecSet PrevHash decode failed %s", err)
	}
	copy(a.PrevHash[:], tmp[:32])
	err = DecodeInt(buf, &a.MemCnt)
	if err != nil {
		return fmt.Errorf("TxDecSet MemCnt decode failed %s", err)
	}
	err = DecodeInt(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxDecSet TxCnt decode failed %s", err)
	}
	for i := uint32(0); i < a.MemCnt; i++ {
		var tmp1 TxDecision
		err = tmp1.Decode(buf)
		if err != nil {
			return fmt.Errorf("TxDecSet MemDecision decode failed %s", err)
		}
		a.MemD = append(a.MemD, tmp1)
	}
	for i := uint32(0); i < a.TxCnt; i++ {
		err = DecodeByteL(buf, &tmp, 32)
		if err != nil {
			return fmt.Errorf("TxDecSet TxArray decode failed %s", err)
		}
		var tmp1 [32]byte
		copy(tmp1[:], tmp[:32])
		a.TxArray = append(a.TxArray, tmp1)
	}
	err = a.Sig.DataToSign(buf)
	if err != nil {
		return fmt.Errorf("TxDecSet Signature decode failed %s", err)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxDecSet decode failed: With extra bits %s", err)
	}
	return nil
}
