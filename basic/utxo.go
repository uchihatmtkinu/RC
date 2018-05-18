package basic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/uchihatmtkinu/RC/crypto"
)

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
func OutToData(a *OutType) []byte {
	/*buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.Value)
	tmp := append(buf.Bytes(), a.Address[:]...)
	return tmp*/
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		fmt.Println(err)
	}

	return result.Bytes()
}

//DataToOut converts bytes into output address data
func DataToOut(data []byte) OutType {
	/*var tmp OutType
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &tmp.Value)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	err = binary.Read(buf, binary.LittleEndian, &tmp.Address)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	return tmp*/
	var tmp OutType
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tmp)
	if err != nil {
		fmt.Println(err)
	}
	return tmp
}

//InToData converts the input address data into bytes
func InToData(a *InType) []byte {
	/*tmp := []byte{}
	tmp1 := a.PrevTx[:]
	EncodeByteL(&tmp, &tmp1, 32)
	EncodeInt(&tmp, a.Index)
	EncodeDoubleBig(&tmp, a.SignR, a.SignS)
	EncodeDoubleBig(&tmp, a.Puk.X, a.Puk.Y)
	//fmt.Println(len(a.PrevTx), len(buf.Bytes()), lenX, lenY, lenPX, lenPY, len(tmp))
	return tmp*/
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err != nil {
		fmt.Println(err)
	}

	return result.Bytes()
}

//DataToIn converts bytes into input address data
func DataToIn(data []byte) InType {
	/*var tmp InType
	var tmp1 []byte
	DecodeByteL(&data, &tmp1, 32)
	copy(tmp.PrevTx[:], tmp1[:32])
	DecodeInt(&data, &tmp.Index)
	tmp.SignR = new(big.Int)
	tmp.SignS = new(big.Int)
	DecodeDoubleBig(&data, tmp.SignR, tmp.SignS)
	tmp.Puk.Curve = elliptic.P256()
	tmp.Puk.X = new(big.Int)
	tmp.Puk.Y = new(big.Int)
	DecodeDoubleBig(&data, tmp.Puk.X, tmp.Puk.Y)
	return tmp*/
	var tmp InType
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tmp)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(tmp.Index, tmp.SignR)
	return tmp
}
