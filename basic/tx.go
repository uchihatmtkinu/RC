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
	var tmp TransactionPure
	tmp.Timestamp = a.Timestamp
	tmp.TxinCnt = a.TxinCnt
	tmp.TxoutCnt = a.TxoutCnt
	tmp.Kind = a.Kind
	tmp.Locktime = a.Locktime
	tmp.Out = a.Out
	tmp.In = nil
	for i := uint32(0); i < a.TxinCnt; i++ {
		var xxx InTypePure
		xxx.Acc = a.In[i].Acc
		xxx.Index = a.In[i].Index
		xxx.PreTx = a.In[i].PrevTx
		tmp.In = append(tmp.In, xxx)
	}
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		fmt.Println(err)
	}
	h := result.Bytes()
	tmpHash := new([32]byte)
	DoubleHash256(&h, tmpHash)
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

//TxToData converts the transaction into bytes
func TxToData(tx *Transaction) []byte {
	/*tmp := []byte{}
	EncodeInt(&tmp, tx.Timestamp)
	EncodeInt(&tmp, tx.TxinCnt)
	EncodeInt(&tmp, tx.TxoutCnt)
	EncodeInt(&tmp, tx.Kind)
	EncodeInt(&tmp, tx.Locktime)
	for i := uint32(0); i < tx.TxinCnt; i++ {
		xxx := InToData(&tx.In[i])
		EncodeByte(&tmp, &xxx)
	}
	for i := uint32(0); i < tx.TxoutCnt; i++ {
		xxx := OutToData(&tx.Out[i])
		tmp = append(tmp, xxx...)
	}*/
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(*tx)
	if err != nil {
		fmt.Println(err)
	}
	return result.Bytes()
}

//DataToTx decodes the packets into transaction format
func DataToTx(data *[]byte) Transaction {
	var tmp Transaction
	/*buf := *data
	err := DecodeInt(&buf, &tmp.Timestamp)
	if err != nil {
		fmt.Println("TX timestamp Read failed:", err)
	}
	err = DecodeInt(&buf, &tmp.TxinCnt)
	if err != nil {
		fmt.Println("TX TxinCnt Read failed:", err)
	}
	err = DecodeInt(&buf, &tmp.TxoutCnt)
	if err != nil {
		fmt.Println("TX TxoutCnt Read failed:", err)
	}
	err = DecodeInt(&buf, &tmp.Kind)
	if err != nil {
		fmt.Println("TX Kind Read failed:", err)
	}
	err = DecodeInt(&buf, &tmp.Locktime)
	if err != nil {
		fmt.Println("TX Locktime Readfailed:", err)
	}
	for i := uint32(0); i < tmp.TxinCnt; i++ {
		var tmpArray *[]byte
		err = DecodeByte(&buf, tmpArray)
		if err != nil {
			fmt.Println("Input Address Readfailed:", err)
		}
		tmpIn := DataToIn(*tmpArray)
		tmp.In = append(tmp.In, tmpIn)
	}
	for i := uint32(0); i < tmp.TxoutCnt; i++ {
		var tmpArray *[]byte
		err = DecodeByteL(&buf, tmpArray, 36)
		if err != nil {
			fmt.Println("Output Address Readfailed:", err)
		}
		tmpOut := DataToOut(*tmpArray)
		tmp.Out = append(tmp.Out, tmpOut)
	}*/
	decoder := gob.NewDecoder(bytes.NewReader(*data))
	err := decoder.Decode(&tmp)
	if err != nil {
		fmt.Println(err)
	}
	return tmp
}

//MakeTx implements the method to create a new transaction
func MakeTx(a *[]InType, b *[]OutType, out *Transaction, kind int) error {
	if out == nil {
		return fmt.Errorf("Basic.MakeTx, null transaction")
	}
	out.Timestamp = time.Now().Unix()
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
