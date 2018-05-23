package basic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
)

//Set initiates the TxDecision given the TxList and the account
func (a *TxDecision) Set(ID [32]byte, b *TxList) error {
	a.TxCnt = 0
	a.HashID = b.HashID
	a.ID = ID
	return nil
}

//Add adds one decision given the result
func (a *TxDecision) Add(x bool) error {
	tmpNum := a.TxCnt % 8
	a.TxCnt++
	if tmpNum == 0 {
		a.Decision = append(a.Decision, byte(0))
	}
	tmp := len(a.Decision)
	if x {
		a.Decision[tmp] += 1 << tmpNum
	}
	return nil
}

//Sign signs the TxDecision
func (a *TxDecision) Sign(prk *ecdsa.PrivateKey) {
	var tmp []byte
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, a.ID[:]...)
	tmpHash := sha256.Sum256(tmp)
	a.Sig.Sign(tmpHash[:], prk)
}

//Verify the signature using public key
func (a *TxDecision) Verify(puk *ecdsa.PublicKey) bool {
	var tmp []byte
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, a.ID[:]...)
	tmpHash := sha256.Sum256(tmp)
	return a.Sig.Verify(tmpHash[:], puk)
}

//Encode encodes the TxDecision into []byte
func (a *TxDecision) Encode(tmp *[]byte) error {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err == nil {
		*tmp = result.Bytes()
	}
	return err
}

// Decode decodes the []byte into TxDecision
func (a *TxDecision) Decode(buf *[]byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(*buf))
	err := decoder.Decode(a)
	return err
}
