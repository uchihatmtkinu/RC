package basic

import (
	"crypto/ecdsa"
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
		tmp = append(tmp, a.In[i].Byte()...)
	}
	for i := uint32(0); i < a.TxoutCnt; i++ {
		tmp = append(tmp, a.Out[i].OutToData()...)
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

//TxToData converts the transaction into bytes
func (a *Transaction) TxToData() []byte {
	tmp := []byte{}
	EncodeInt(&tmp, a.Timestamp)
	EncodeInt(&tmp, a.TxinCnt)
	EncodeInt(&tmp, a.TxoutCnt)
	EncodeInt(&tmp, a.Kind)
	EncodeInt(&tmp, a.Locktime)
	EncodeByteL(&tmp, a.Hash[:], 32)
	for i := uint32(0); i < a.TxinCnt; i++ {
		xxx := a.In[i].InToData()
		//EncodeByte(&tmp, &xxx)
		tmp = append(tmp, xxx...)
	}
	for i := uint32(0); i < a.TxoutCnt; i++ {
		xxx := a.Out[i].OutToData()
		tmp = append(tmp, xxx...)
	}
	return tmp
}

//DataToTx decodes the packets into transaction format
func (a *Transaction) DataToTx(buf *[]byte) error {
	//buf := *data

	err := DecodeInt(buf, &a.Timestamp)
	if err != nil {
		return fmt.Errorf("TX timestamp Read failed %s", err)
	}
	err = DecodeInt(buf, &a.TxinCnt)
	if err != nil {
		return fmt.Errorf("TX TxinCnt Read failed %s", err)
	}
	err = DecodeInt(buf, &a.TxoutCnt)
	if err != nil {
		return fmt.Errorf("TX TxoutCnt Read failed %s", err)
	}
	err = DecodeInt(buf, &a.Kind)
	if err != nil {
		return fmt.Errorf("TX Kind Read failed %s", err)
	}
	err = DecodeInt(buf, &a.Locktime)
	if err != nil {
		return fmt.Errorf("TX Locktime Readfailed %s", err)
	}
	var tmp1 []byte
	err = DecodeByteL(buf, &tmp1, 32)
	if err != nil {
		return fmt.Errorf("TX hash Readfailed %s", err)
	}
	copy(a.Hash[:], tmp1[:32])
	for i := uint32(0); i < a.TxinCnt; i++ {
		//var tmpArray *[]byte
		var tmpIn InType
		err = tmpIn.DataToIn(buf)
		if err != nil {
			return fmt.Errorf("Input Address Readfailed %s", err)
		}
		a.In = append(a.In, tmpIn)
	}

	for i := uint32(0); i < a.TxoutCnt; i++ {
		//var tmpArray *[]byte
		var tmpOut OutType
		err = tmpOut.DataToOut(buf)
		if err != nil {
			return fmt.Errorf("Output Address Readfailed %s", err)
		}
		a.Out = append(a.Out, tmpOut)
	}
	return nil
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
