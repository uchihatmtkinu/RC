package basic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
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
func (a *TxDecSet) Encode(tmp *[]byte) error {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err == nil {
		*tmp = result.Bytes()
	}
	return err
}

//Decode encode the TxDecSet into []byte
func (a *TxDecSet) Decode(buf *[]byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(*buf))
	err := decoder.Decode(a)
	return err
}
