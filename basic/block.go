package basic

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"
)

//HashTxBlock generates the 32bits hash of one Tx block
func HashTxBlock(a *TxBlock, b *[32]byte) {
	tmp1 := []byte{}
	tmp1 = append(tmp1, a.PrevHash[:]...)
	tmp1 = append(tmp1, a.MerkleRoot[:]...)
	var tmp2 *[]byte
	EncodeInt(tmp2, a.Timestamp)
	tmp1 = append(tmp1, *tmp2...)
	DoubleHash256(&tmp1, b)
}

//GenMerkTree generates the merkleroot tree given the transactions
func GenMerkTree(d *[]RawTransaction, out *[32]byte) error {
	if len(*d) == 1 {
		tmp := (*d)[0].Hash[:]
		DoubleHash256(&tmp, out)
	} else {
		l := len(*d)
		d1 := (*d)[:l/2]
		d2 := (*d)[l/2:]
		var out1, out2 [32]byte
		GenMerkTree(&d1, &out1)
		GenMerkTree(&d2, &out2)
		tmp := append(out1[:], out2[:]...)
		DoubleHash256(&tmp, out)
	}
	return nil
}

//VerifyTxBlock verify the signature of transaction a
func VerifyTxBlock(a *TxBlock) (bool, error) {

	var tmp [32]byte
	HashTxBlock(a, &tmp)
	//Verify the hash, the cnt of in and out address
	if tmp != a.HashID || a.TxCnt != uint32(len(a.TxArray)) {
		return false, fmt.Errorf("VerifyTxBlock Invalid parameter")
	}
	if !ecdsa.Verify(&a.Prk, a.HashID[:], a.SignR, a.SignS) {
		return false, fmt.Errorf("VerifyTxBlock Invalid signature")
	}
	return false, fmt.Errorf("VerifyTx.Invalid transaction type")
}

//MakeTxBlock creates the transaction blocks given verified transactions
func MakeTxBlock(a *[]RawTransaction, preHash [32]byte, prk *ecdsa.PrivateKey, h uint32, out *TxBlock) error {
	if out == nil {
		return fmt.Errorf("Basic.MakeTxBlock, null block")
	}

	out.PrevHash = preHash
	out.Timestamp = time.Now().Unix()
	out.TxCnt = uint32(len(*a))
	out.Height = h
	out.TxArray = []RawTransaction{}
	for i := 0; i < int(out.TxCnt); i++ {
		out.TxArray = append(out.TxArray, (*a)[i])
	}
	GenMerkTree(&out.TxArray, &out.MerkleRoot)

	HashTxBlock(out, &out.HashID)
	out.Prk = prk.PublicKey
	out.SignR, out.SignS, _ = ecdsa.Sign(rand.Reader, prk, out.HashID[:])

	return nil
}

//BlockToData converts the block data into bytes
func BlockToData(b *TxBlock) []byte {
	tmp := []byte{}
	tmp1 := b.PrevHash[:]
	EncodeByteL(&tmp, &tmp1, 32)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, b.TxCnt)
	binary.Write(buf, binary.LittleEndian, b.Timestamp)
	binary.Write(buf, binary.LittleEndian, b.Height)
	tmp = append(tmp, buf.Bytes()...)
	tmp1 = b.HashID[:]
	EncodeByteL(&tmp, &tmp1, 32)
	tmp1 = b.MerkleRoot[:]
	EncodeByteL(&tmp, &tmp1, 32)
	EncodeDoubleBig(&tmp, b.SignR, b.SignS)
	EncodeDoubleBig(&tmp, b.Prk.X, b.Prk.Y)
	for i := uint32(0); i < b.TxCnt; i++ {
		tmpTxData := TxToData(&b.TxArray[i])
		EncodeByte(&tmp, &tmpTxData)
	}

	return tmp
}

//DataToBlock converts bytes into block data
func DataToBlock(data *[]byte) (TxBlock, error) {
	buf := *data
	var b TxBlock
	var tmp []byte
	err := DecodeByteL(&buf, &tmp, 32)
	if err != nil {
		return b, fmt.Errorf("DataToBlock PrevHash failed %s", err)
	}
	copy(b.PrevHash[:], tmp[:32])
	err = DecodeInt(&buf, &b.TxCnt)
	if err != nil {
		return b, fmt.Errorf("DataToBlock TxCnt failed %s", err)
	}
	err = DecodeInt(&buf, &b.Timestamp)
	if err != nil {
		return b, fmt.Errorf("DataToBlock Timestamp failed %s", err)
	}
	err = DecodeInt(&buf, &b.Height)
	if err != nil {
		return b, fmt.Errorf("DataToBlock Height failed %s", err)
	}
	err = DecodeByteL(&buf, &tmp, 32)
	if err != nil {
		return b, fmt.Errorf("DataToBlock HashID failed %s", err)
	}
	copy(b.HashID[:], tmp[:32])
	err = DecodeByteL(&buf, &tmp, 32)
	if err != nil {
		return b, fmt.Errorf("DataToBlock MerkleRoot failed %s", err)
	}
	copy(b.MerkleRoot[:], tmp[:32])
	big1 := new(big.Int)
	big2 := new(big.Int)
	err = DecodeDoubleBig(&buf, big1, big2)
	if err != nil {
		return b, fmt.Errorf("DataToBlock Signature failed %s", err)
	}
	*b.SignR = *big1
	*b.SignS = *big2
	err = DecodeDoubleBig(&buf, big1, big2)
	if err != nil {
		return b, fmt.Errorf("DataToBlock Public Key failed %s", err)
	}
	b.Prk.Curve = elliptic.P256()
	*b.Prk.X = *big1
	*b.Prk.Y = *big2
	for i := uint32(0); i < b.TxCnt; i++ {
		var tmpBuf []byte
		err = DecodeByte(&buf, &tmpBuf)
		if err != nil {
			return b, fmt.Errorf("DataToBlock decode Tx failed %s", err)
		}
		tmpTx := DataToTx(&tmpBuf)
		b.TxArray = append(b.TxArray, tmpTx)
	}
	//DecodeByte(&buf, &b.Signature)
	return b, nil
}
