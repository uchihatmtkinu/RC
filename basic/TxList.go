package basic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
)

//Hash returns the ID of the TxList
func (a *TxList) Hash() [32]byte {
	tmp := a.TxArray[0].Hash[:]
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp = append(tmp, a.TxArray[i].Hash[:]...)
	}
	return sha256.Sum256(tmp)
}

//Sign signs the TxList with the leader's private key
func (a *TxList) Sign(prk *ecdsa.PrivateKey) {
	a.Sig.Sign(a.HashID[:], prk)
}

//Verify verify the signature
func (a *TxList) Verify(puk *ecdsa.PublicKey) bool {
	return a.Sig.Verify(a.HashID[:], puk)
}

//Set init an instance of TxList given those parameters
func (a *TxList) Set(ID [32]byte, prevH [32]byte) {
	a.ID = ID
	a.PrevHash = prevH
	a.TxCnt = 0
	a.TxArray = nil
}

//AddTx adds the tx into transaction list
func (a *TxList) AddTx(tx *Transaction) {
	a.TxCnt++
	a.TxArray = append(a.TxArray, *tx)
}

//Encode returns the byte of a TxList
func (a *TxList) Encode(tmp *[]byte) error {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err == nil {
		*tmp = result.Bytes()
	}
	return err
}

//Decode decodes the TxList with []byte
func (a *TxList) Decode(buf *[]byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(*buf))
	err := decoder.Decode(a)
	return err
}
