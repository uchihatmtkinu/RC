package basic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/uchihatmtkinu/RC/crypto"
)

//Init initial the big.Int parameters
func (a *InType) Init() {
	a.PukX = new(big.Int)
	a.PukY = new(big.Int)
	a.SignR = new(big.Int)
	a.SignS = new(big.Int)
}

//Byte return the []byte of the input address used for hash
func (a *InType) Byte(b *[]byte) {
	EncodeByteL(b, a.PrevTx[:], 32)
	EncodeInt(b, a.Index)
	EncodeInt(b, a.Acc)
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
	if !ecdsa.Verify(&tmpPuk, h[:], a.SignR, a.SignS) {
		fmt.Println("UTXO.VerifyIn signature incorrect")
		return false
	}
	return true
}

//SignTxIn make the signature given the transaction
func (a *InType) SignTxIn(prk *ecdsa.PrivateKey, h [32]byte) {
	a.PukX.Set(prk.PublicKey.X)
	a.PukY.Set(prk.PublicKey.Y)
	a.SignR = new(big.Int)
	a.SignS = new(big.Int)
	a.SignR, a.SignS, _ = ecdsa.Sign(rand.Reader, prk, h[:])
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
	var tmp1 []byte
	err := DecodeInt(data, &a.Value)
	if err != nil {
		return fmt.Errorf("Out Value read failed")
	}
	err = DecodeByteL(data, &tmp1, 32)
	if err != nil {
		return fmt.Errorf("Out Address read failed")
	}
	copy(a.Address[:], tmp1[:32])
	return nil
}

//InToData converts the input address data into bytes
func (a *InType) InToData(b *[]byte) {
	EncodeByteL(b, a.PrevTx[:], 32)
	EncodeInt(b, a.Index)
	EncodeDoubleBig(b, a.SignR, a.SignS)
	EncodeDoubleBig(b, a.PukX, a.PukY)
	EncodeInt(b, a.Acc)
	//fmt.Println(len(a.PrevTx), len(buf.Bytes()), lenX, lenY, lenPX, lenPY, len(tmp))
}

//DataToIn converts bytes into input address data
func (a *InType) DataToIn(data *[]byte) error {
	var tmp1 []byte
	a.Init()
	err := DecodeByteL(data, &tmp1, 32)
	if err != nil {
		return fmt.Errorf("Input PrevHash read failed %s", err)
	}
	copy(a.PrevTx[:], tmp1[:32])
	err = DecodeInt(data, &a.Index)
	if err != nil {
		return fmt.Errorf("Input Index read failed %s", err)
	}
	err = DecodeDoubleBig(data, a.SignR, a.SignS)
	if err != nil {
		return fmt.Errorf("Input Signature read failed %s", err)
	}
	err = DecodeDoubleBig(data, a.PukX, a.PukY)
	if err != nil {
		return fmt.Errorf("Input Publick Key read failed %s", err)
	}
	err = DecodeInt(data, &a.Acc)
	if err != nil {
		return fmt.Errorf("Input Account type read failed %s", err)
	}
	return nil

}
