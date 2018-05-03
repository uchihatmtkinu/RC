package basic

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
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

//MakeTxBlock creates the transaction blocks given verified transactions
func MakeTxBlock(a *[]RawTransaction, preHash [32]byte, prk *ecdsa.PrivateKey, h uint32) *TxBlock {
	tmp := new(TxBlock)
	tmp.PrevHash = preHash
	tmp.Timestamp = time.Now().Unix()
	tmp.TxCnt = uint32(len(*a))
	tmp.Height = h
	HashTxBlock(tmp, &tmp.HashID)
	tmp.TxArray = []RawTransaction{}
	for i := 0; i < int(tmp.TxCnt); i++ {
		tmp.TxArray = append(tmp.TxArray, (*a)[i])
	}

	//tmp.Signature, _ = ecdsa.Sign(rand.Reader, prk, tmp.HashID[:])
	return tmp
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

	for i := uint32(0); i < b.TxCnt; i++ {
		tmpTxData := TxToData(&b.TxArray[i])
		EncodeByte(&tmp, &tmpTxData)
	}

	//EncodeByte(&tmp, &b.Signature)
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
