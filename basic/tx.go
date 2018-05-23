package basic

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"fmt"
	"time"
)

//HashTx is come out the hash
func (a *Transaction) HashTx() [32]byte {
	var tmp []byte
	EncodeInt(&tmp, a.Timestamp)
	EncodeInt(&tmp, a.TxinCnt)
	EncodeInt(&tmp, a.TxoutCnt)
	EncodeInt(&tmp, a.Kind)
	EncodeInt(&tmp, a.Locktime)
	for i := uint32(0); i < a.TxinCnt; i++ {
		a.In[i].Byte(&tmp)
	}
	for i := uint32(0); i < a.TxoutCnt; i++ {
		a.Out[i].OutToData(&tmp)
	}
	tmpHash := new([32]byte)
	DoubleHash256(&tmp, tmpHash)
	return *tmpHash
}

//SignTx sign the ith in-address with the private key
func (a *Transaction) SignTx(i uint32, prk *ecdsa.PrivateKey) error {
	if a.TxinCnt <= i {
		return fmt.Errorf("Tx.SignTx in-address outrange")
	}
	a.In[i].SignTxIn(prk, a.Hash)
	return nil
}

//VerifyTx sign the ith in-address with the private key
func (a *Transaction) VerifyTx(i uint32, b *OutType) bool {
	if a.TxinCnt <= i {
		fmt.Println("Tx.VerifyTx in-address outrange")
		return false
	}

	return a.In[i].VerifyIn(b, a.Hash)
}

//Encode converts the transaction into bytes
func (a *Transaction) Encode(tmp *[]byte) error {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err == nil {
		*tmp = result.Bytes()
	}
	return err
}

//Decode decodes the packets into transaction format
func (a *Transaction) Decode(buf *[]byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(*buf))
	err := decoder.Decode(a)
	return err
}

//MakeTx implements the method to create a new transaction
func MakeTx(a *[]InType, b *[]OutType, out *Transaction, kind int) error {
	if out == nil {
		return fmt.Errorf("Basic.MakeTx, null transaction")
	}
	out.Timestamp = uint64(time.Now().Unix())
	out.TxinCnt = uint32(len(*a))
	out.TxoutCnt = uint32(len(*b))
	out.Kind = uint32(kind)
	out.In = []InType{}
	for i := 0; i < int(out.TxinCnt); i++ {
		out.In = append(out.In, (*a)[i])
	}
	out.Out = []OutType{}
	for i := 0; i < int(out.TxoutCnt); i++ {
		out.Out = append(out.Out, (*b)[i])
	}
	out.Hash = out.HashTx()
	return nil
}

//AddIn increases one input of transaction a
func (a *Transaction) AddIn(b InType) {
	a.TxinCnt++
	a.In = append(a.In, b)
}

//AddOut increases one output of transaction a
func (a *Transaction) AddOut(b OutType) {
	a.TxoutCnt++
	a.Out = append(a.Out, b)
}
