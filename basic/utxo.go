package basic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/uchihatmtkinu/RC/cryptonew"
)

//Init initial the big.Int parameters
func (a *InType) Init() {
	a.PukX = new(big.Int)
	a.PukY = new(big.Int)
}

//Acc returns whether the input address is an account or utxo
func (a *InType) Acc() bool {
	return cryptonew.Verify(a.Puk(), a.PrevTx)
}

//ShardIndex returns the target shard of the input address
func (a *InType) ShardIndex() uint32 {
	tmp := cryptonew.GenerateAddr(a.Puk())
	return uint32(tmp[0]) % ShardCnt
}

//Byte return the []byte of the input address used for hash
func (a *InType) Byte(b *[]byte) {
	EncodeByteL(b, a.PrevTx[:], 32)
	EncodeInt(b, a.Index)
}

//Puk returns the public key
func (a *InType) Puk() ecdsa.PublicKey {
	var tmp ecdsa.PublicKey
	tmp.Curve = elliptic.P256()
	tmp.X = a.PukX
	tmp.Y = a.PukY
	return tmp
}

//VerifyIn using the UTXO to verify the in address
func (a *InType) VerifyIn(b *OutType, h [32]byte) bool {
	if !cryptonew.Verify(a.Puk(), b.Address) {
		fmt.Println("UTXO.VerifyIn address doesn't match")
		return false
	}
	tmpPuk := a.Puk()
	if !a.Sig.Verify(h[:], &tmpPuk) {
		fmt.Println("UTXO.VerifyIn signature incorrect")
		return false
	}
	return true
}

//SignTxIn make the signature given the transaction
func (a *InType) SignTxIn(prk *ecdsa.PrivateKey, h [32]byte) {
	a.PukX.Set(prk.PublicKey.X)
	a.PukY.Set(prk.PublicKey.Y)
	a.Sig.Sign(h[:], prk)
}

//OutToData converts the output address data into bytes
func (a *OutType) OutToData(b *[]byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.Value)
	*b = append(*b, buf.Bytes()...)
	*b = append(*b, a.Address[:]...)
}

//DataToOut converts bytes into output address data
func (a *OutType) DataToOut(data *[]byte) error {
	tmp1 := make([]byte, 0, 32)
	err := DecodeInt(data, &a.Value)
	if err != nil {
		return fmt.Errorf("Out Value read failed: %s", err)
	}
	err = DecodeByteL(data, &tmp1, 32)
	if err != nil {
		return fmt.Errorf("Out Address read failed: %s", err)
	}
	copy(a.Address[:], tmp1[:32])
	return nil
}

//InToData converts the input address data into bytes
func (a *InType) InToData(b *[]byte) {
	EncodeByteL(b, a.PrevTx[:], 32)
	EncodeInt(b, a.Index)
	a.Sig.SignToData(b)
	EncodeDoubleBig(b, a.PukX, a.PukY)
	//fmt.Println(len(a.PrevTx), len(buf.Bytes()), lenX, lenY, lenPX, lenPY, len(tmp))
}

//DataToIn converts bytes into input address data
func (a *InType) DataToIn(data *[]byte) error {
	tmp1 := make([]byte, 0, 32)
	a.Init()
	err := DecodeByteL(data, &tmp1, 32)
	if err != nil {
		return fmt.Errorf("Input PrevHash read failed: %s", err)
	}
	copy(a.PrevTx[:], tmp1[:32])
	err = DecodeInt(data, &a.Index)
	if err != nil {
		return fmt.Errorf("Input Index read failed: %s", err)
	}
	err = a.Sig.DataToSign(data)
	if err != nil {
		return fmt.Errorf("Input Signature read failed: %s", err)
	}
	err = DecodeDoubleBig(data, a.PukX, a.PukY)
	if err != nil {
		return fmt.Errorf("Input Publick Key read failed: %s", err)
	}
	return nil

}
