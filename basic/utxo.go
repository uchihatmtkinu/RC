package basic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/uchihatmtkinu/RC/crypto"
)

//SearchUTXO find the specific out address given the hash of the transaction and index
func SearchUTXO(index uint32, h *[32]byte, out *OutType, db *[]TxDB) uint32 {
	//We implement O(n) search at current, future change it into O(nlogn)
	for i := range *db {
		if (*db)[i].Data.Hash == *h {
			if (*db)[i].Data.TxoutCnt > uint32(index) {
				out = &(*db)[i].Data.Out[index]
				return (*db)[i].Used[index]
			}
			return 20
		}
	}
	out = nil
	return 10
}

//Prk returns the public key
func (b *InType) Prk() ecdsa.PublicKey {
	var tmp ecdsa.PublicKey
	tmp.Curve = elliptic.P256()
	tmp.X = b.PrkX
	tmp.Y = b.PrkY
	return tmp
}

//VerifyIn using the UTXO to verify the in address
func (b *InType) VerifyIn(a *OutType, h [32]byte) bool {
	if !cryptonew.Verify(b.Prk(), a.Address) {
		return false
	}
	tmpx := b.Prk()
	if !ecdsa.Verify(&tmpx, h[:], b.SignR, b.SignS) {
		return false
	}
	return true
}

//VerifyTxIn implements the function to verify the signature to use the current UTXO
func VerifyTxIn(a *InType, out uint32, db *[]TxDB) bool {
	var b *OutType
	err := SearchUTXO(a.Index, &a.PrevTx, b, db)
	if err != 0 {
		return false
	}

	if !cryptonew.Verify(a.Prk(), b.Address) {
		return false
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.Index)
	tmp := append(a.PrevTx[:], buf.Bytes()...)
	tmpHash := new([32]byte)
	DoubleHash256(&tmp, tmpHash)
	tmpx := a.Prk()
	if !ecdsa.Verify(&tmpx, tmpHash[:], a.SignR, a.SignS) {
		return false
	}
	return true
}

//SignTxIn make the signature given the transaction
func SignTxIn(a *InType, prk *ecdsa.PrivateKey, h [32]byte) {
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
