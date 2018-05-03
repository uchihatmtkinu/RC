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

//SearchUTXO find the specific out address given the hash of the transaction and index
func SearchUTXO(index uint32, h *[32]byte, out *OutType, db *[]TxDB) int {
	//We implement O(n) search at current, future change it into O(nlogn)
	for i := range *db {
		if (*db)[i].data.Hash == *h {
			if (*db)[i].data.TxoutCnt > uint32(index) {
				out = &(*db)[i].data.Out[index]
				return (*db)[i].used[index]
			}
			return -2
		}
	}
	out = nil
	return -1
}

//VerifyTxIn implements the function to verify the signature to use the current UTXO
func VerifyTxIn(a *InType, out uint32, db *[]TxDB) bool {
	var b *OutType
	err := SearchUTXO(a.Index, &a.PrevTx, b, db)
	if err != 0 {
		return false
	}
	if !cryptonew.Verify(a.Puk, b.Address) {
		return false
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.Index)
	tmp := append(a.PrevTx[:], buf.Bytes()...)
	tmpHash := new([32]byte)
	DoubleHash256(&tmp, tmpHash)
	if !ecdsa.Verify(&a.Puk, tmpHash[:], a.SignR, a.SignS) {
		return false
	}
	return true
}

//SignTxIn make the signature given the transaction
func SignTxIn(a *InType, prk *ecdsa.PrivateKey) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.Index)
	tmpHash := new([32]byte)
	tmp := append(a.PrevTx[:], buf.Bytes()...)
	DoubleHash256(&tmp, tmpHash)
	a.SignR = new(big.Int)
	a.SignS = new(big.Int)
	a.SignR, a.SignS, _ = ecdsa.Sign(rand.Reader, prk, tmpHash[:])
}

//OutToData converts the output address data into bytes
func OutToData(a *OutType) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, a.Value)
	tmp := append(buf.Bytes(), a.Address[:]...)
	return tmp
}

//DataToOut converts bytes into output address data
func DataToOut(data []byte) OutType {
	var tmp OutType
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &tmp.Value)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	err = binary.Read(buf, binary.LittleEndian, &tmp.Address)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	return tmp
}

//InToData converts the input address data into bytes
func InToData(a *InType) []byte {
	tmp := []byte{}
	tmp1 := a.PrevTx[:]
	EncodeByteL(&tmp, &tmp1, 32)
	EncodeInt(&tmp, a.Index)
	EncodeDoubleBig(&tmp, a.SignR, a.SignS)
	EncodeDoubleBig(&tmp, a.Puk.X, a.Puk.Y)
	//fmt.Println(len(a.PrevTx), len(buf.Bytes()), lenX, lenY, lenPX, lenPY, len(tmp))
	return tmp
}

//DataToIn converts bytes into input address data
func DataToIn(data []byte) InType {
	var tmp InType
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
	return tmp
}
