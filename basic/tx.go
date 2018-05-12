package basic

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"
)

//TxToData converts the transaction into bytes
func TxToData(tx *RawTransaction) []byte {
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
func DataToTx(data *[]byte) RawTransaction {
	var tmp RawTransaction
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
func MakeTx(a *[]InType, b *[]OutType, out *RawTransaction, kind int) error {
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
	HashTx(out, &out.Hash)
	return nil
}
