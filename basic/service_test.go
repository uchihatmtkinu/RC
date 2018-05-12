package basic

import (
	"crypto/sha256"
	"math/big"
	"testing"
)

func TestOutToData(t *testing.T) {
	var tmp OutType
	xxx := sha256.Sum256([]byte("test2"))
	tmp.Address = xxx
	tmp.Value = 15
	yyy := DataToOut(OutToData(&tmp))
	if yyy.Address != tmp.Address {
		t.Error(`Address is wrong`)
	}
	if yyy.Value != tmp.Value {
		t.Error(`Value is wrong`)
	}
}

func TestInToData(t *testing.T) {
	var tmp InType
	tmp.PrevTx = sha256.Sum256([]byte("test2"))
	tmp.Index = 10
	tmp.SignR = new(big.Int)
	tmp.SignS = new(big.Int)
	tmp.PrkX = new(big.Int)
	tmp.PrkY = new(big.Int)
	tmp.SignR.SetString("123123123123", 10)
	tmp.SignS.SetString("12345", 10)
	tmp.PrkX.SetString("1234567890", 10)
	tmp.PrkY.SetString("123123123123", 10)
	xxx := DataToIn(InToData(&tmp))
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
	if tmp.PrkX.Cmp(xxx.PrkX) != 0 {
		t.Error(`Puk X is wrong`)
	}
	if tmp.PrkY.Cmp(xxx.PrkY) != 0 {
		t.Error(`Puk Y is wrong`)
	}
}
