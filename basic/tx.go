package basic

import (
	"crypto/sha256"
	"fmt"
	"time"
)

//TxToData converts the transaction into bytes
func TxToData(tx *RawTransaction) []byte {
	tmp := []byte{}
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
	}
	return tmp
}

//DataToTx decodes the packets into transaction format
func DataToTx(data *[]byte) RawTransaction {
	var tmp RawTransaction
	buf := *data
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
	}
	return tmp
}

//HashTx generates the 32bits hash of one transaction
func HashTx(a *RawTransaction, b *[32]byte) {
	tmp1 := TxToData(a)
	DoubleHash256(&tmp1, b)
}

//VerifyTx verify the signature of transaction a
func VerifyTx(a *RawTransaction, db *[]TxDB) (bool, error) {

	tmp := sha256.Sum256(TxToData(a))
	//Verify the hash, the cnt of in and out address
	if tmp != a.Hash || a.TxinCnt != uint32(len(a.In)) || a.TxoutCnt != uint32(len(a.Out)) {
		return false, fmt.Errorf("Invalid parameter")
	}
	//Verify when it is a normal transaction
	if a.Kind == 0 {
		var value uint32
		var tmpInt uint32
		for i := uint32(0); i < a.TxinCnt; i++ {
			if !VerifyTxIn(&a.In[i], tmpInt, db) {
				return false, fmt.Errorf("VerifyTx.Invalid UTXO of %d", &i)
			}
			value += tmpInt
		}
		total := value
		for i := uint32(0); i < a.TxoutCnt; i++ {
			value -= a.Out[i].Value
		}
		if value*100 < total {
			return false, fmt.Errorf("VerifyTx.Invalid outcome value")
		}
		return true, nil
	} else if a.Kind == 1 { //Verify when it is a transafer transaction
		if a.TxoutCnt != 1 {
			return false, fmt.Errorf("VerifyTx.The out address should be 1")
		}
		for i := uint32(0); i < a.TxinCnt; i++ {
			var b *OutType
			tmp := SearchUTXO(a.In[i].Index, &a.In[i].PrevTx, b, db)
			if tmp != 0 {
				return false, fmt.Errorf("VerifyTx.Invalid out address")
			}
			if b.Address != a.Out[0].Address {
				return false, fmt.Errorf("VerifyTx.Unmatch income address")
			}
		}
		return true, nil
	}
	return false, fmt.Errorf("VerifyTx.Invalid transaction type")
}

//MakeTx implements the method to create a new transaction
func MakeTx(a *[]InType, b *[]OutType, out *RawTransaction, value *[]int, kind int) {
	tmp := new(RawTransaction)
	tmp.Timestamp = time.Now().Unix()
	tmp.TxinCnt = uint32(len(*a))
	tmp.TxoutCnt = uint32(len(*b))
	tmp.Kind = uint32(kind)
	tmp.In = []InType{}
	for i := 0; i < int(tmp.TxinCnt); i++ {
		tmp.In = append(tmp.In, (*a)[i])
	}
	tmp.Out = []OutType{}
	for i := 0; i < int(tmp.TxoutCnt); i++ {
		tmp.Out = append(tmp.Out, (*b)[i])
	}
	HashTx(tmp, &tmp.Hash)
}
