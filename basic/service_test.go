package basic

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"
)

func TestOutToData(t *testing.T) {
	var tmp, yyy OutType
	xxx := sha256.Sum256([]byte("test2"))
	tmp.Address = xxx
	tmp.Value = 15
	tmpOut := tmp.OutToData()

	yyy.DataToOut(&tmpOut)
	if yyy.Address != tmp.Address {
		t.Error(`Address is wrong`)
	}
	if yyy.Value != tmp.Value {
		t.Error(`Value is wrong`)
	}
}

func TestInToData(t *testing.T) {
	var tmp, xxx InType
	tmp.PrevTx = sha256.Sum256([]byte("test2"))
	tmp.Index = 10
	tmp.SignR = new(big.Int)
	tmp.SignS = new(big.Int)
	tmp.PukX = new(big.Int)
	tmp.PukY = new(big.Int)
	tmp.SignR.SetString("123123123123", 10)
	tmp.SignS.SetString("12345", 10)
	tmp.PukX.SetString("1234567890", 10)
	tmp.PukY.SetString("123123123123", 10)
	for i := 1; i < 500; i++ {
		tmpIn := tmp.InToData()
		xxx.DataToIn(&tmpIn)
		if tmp.PrevTx != xxx.PrevTx {
			t.Error(`Prev Hash is wrong`)
		}
		if tmp.Index != xxx.Index {
			t.Error(`Index is wrong`, tmp.Index, xxx.Index)
		}
		if tmp.SignR.Cmp(xxx.SignR) != 0 {
			t.Error(`Sig R is wrong`, tmp.SignR, xxx.SignR)
		}
		if tmp.SignS.Cmp(xxx.SignS) != 0 {
			t.Error(`Sig S is wrong`)
		}
		if tmp.PukX.Cmp(xxx.PukX) != 0 {
			t.Error(`Puk X is wrong`)
		}
		if tmp.PukY.Cmp(xxx.PukY) != 0 {
			t.Error(`Puk Y is wrong`)
		}
	}
}

func TestTxtoData(t *testing.T) {
	numIn := 2
	numOut := 2
	var tmpIn []InType
	var tmpOut []OutType
	for i := 0; i < numIn; i++ {
		var tmpInx InType
		tmpInx.PrevTx = FindByte32(i * 1000)
		tmpInx.Index = uint32(i)
		tmpInx.Init()
		*tmpInx.PukX = FindBigInt((i+1)*1000 + 1)
		*tmpInx.PukY = FindBigInt((i+1)*1000 + 2)
		*tmpInx.SignR = FindBigInt((i+1)*1000 + 3)
		*tmpInx.SignS = FindBigInt((i+1)*1000 + 4)
		tmpInx.Acc = false
		tmpIn = append(tmpIn, tmpInx)
	}

	for i := 0; i < numOut; i++ {
		var tmpOutx OutType
		tmpOutx.Address = FindByte32(i * 2000)
		tmpOutx.Value = uint32(i)
		tmpOut = append(tmpOut, tmpOutx)
	}
	var tmpTx, tmp1 Transaction
	MakeTx(&tmpIn, &tmpOut, &tmpTx, 1)

	tmp := tmpTx.TxToData()
	tmp1.DataToTx(&tmp)
	if tmp1.Timestamp != tmpTx.Timestamp {
		t.Error(`Timestamp is wrong`)
	}
	if tmp1.TxinCnt != tmpTx.TxinCnt {
		t.Error(`TxinCnt is wrong`)
	}
	if tmp1.TxoutCnt != tmpTx.TxoutCnt {
		t.Error(`TxoutCnt is wrong`)
	}
	if tmp1.Kind != tmpTx.Kind {
		t.Error(`Kind is wrong`)
	}
	for i := 0; i < numOut; i++ {
		fmt.Println(tmp1.Out[i].Value, tmpTx.Out[i].Value)
		if tmp1.Out[i].Value != tmpTx.Out[i].Value {
			t.Error(`Output value is wrong`, i)
		}
		if tmp1.Out[i].Address != tmpTx.Out[i].Address {
			t.Error(`Output address is wrong`, i)
		}
	}
	if tmp1.HashTx() != tmpTx.HashTx() {
		t.Error(`Hash is wrong`)
	}

}
