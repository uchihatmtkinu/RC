package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"time"
)

//Hash generates the 32bits hash of one Tx block
func (a *TxBlock) Hash() [32]byte {
	tmp1 := make([]byte, 0, 136)
	tmp1 = append(byteSlice(a.ID), a.PrevHash[:]...)
	tmp1 = append(tmp1, a.HashID[:]...)
	tmp1 = append(tmp1, a.MerkleRoot[:]...)
	EncodeInt(&tmp1, a.Timestamp)
	var b [32]byte
	DoubleHash256(&tmp1, &b)
	return b
}

//GenMerkTree generates the merkleroot tree given the transactions
func GenMerkTree(d *[]Transaction, out *[32]byte) error {
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

//Verify verify the signature of the Txblock
func (a *TxBlock) Verify(puk *ecdsa.PublicKey) (bool, error) {

	tmp := a.Hash()
	//Verify the hash, the cnt of in and out address
	if tmp != a.HashID || a.TxCnt != uint32(len(a.TxArray)) {
		return false, fmt.Errorf("VerifyTxBlock Invalid parameter")
	}
	var tmpHash [32]byte
	GenMerkTree(&a.TxArray, &tmpHash)
	if tmpHash != a.MerkleRoot {
		return false, fmt.Errorf("VerifyTxBlock MerkleRoot Invalid")
	}
	if !a.Sig.Verify(a.HashID[:], puk) {
		return false, fmt.Errorf("VerifyTxBlock Invalid signature")
	}
	return false, fmt.Errorf("VerifyTx.Invalid transaction type")
}

//NewGensisTxBlock is the gensis block
func NewGensisTxBlock() TxBlock {
	var a TxBlock
	a.ID = 0
	a.TxCnt = 0
	a.HashID = sha256.Sum256([]byte(GenesisTxBlock))
	a.Height = 0
	return a
}

//MakeTxBlock creates the transaction blocks given verified transactions
func (a *TxBlock) MakeTxBlock(ID uint32, b *[]Transaction, preHash [32]byte, prk *ecdsa.PrivateKey, h uint32, kind uint32) error {
	a.ID = ID
	a.Kind = kind
	a.PrevHash = preHash
	a.Timestamp = time.Now().Unix()
	a.TxCnt = uint32(len(*b))
	a.Height = h
	for i := 0; i <= len(*b); {
		a.TxArray = append(a.TxArray, HashCut((*b)[i].Hash))
	}

	GenMerkTree(&a.TxArray, &a.MerkleRoot)
	a.HashID = a.Hash()
	a.Sig.Sign(a.HashID[:], prk)
	return nil
}

//Serial outputs a serial of []byte
func (a *TxBlock) Serial() []byte {
	var tmp []byte
	a.Encode(&tmp)
	return tmp
}

//Encode converts the block data into bytes
func (a *TxBlock) Encode(tmp *[]byte) {
	EncodeInt(tmp, a.ID)
	EncodeByteL(tmp, a.PrevHash[:], 32)
	EncodeByteL(tmp, a.HashID[:], 32)
	EncodeByteL(tmp, a.MerkleRoot[:], 32)
	EncodeInt(tmp, a.Timestamp)
	EncodeInt(tmp, a.Height)
	EncodeInt(tmp, a.TxCnt)
	EncodeInt(tmp, a.Kind)
	for i := uint32(0); i < a.TxCnt; i++ {
		EncodeByteL(tmp, a.TxArray[i][:], sHash)
	}
	if a.HashID != sha256.Sum256([]byte(GenesisTxBlock)) {
		a.Sig.SignToData(tmp)
	}
}

//Decode converts bytes into block data
func (a *TxBlock) Decode(buf *[]byte) error {
	tmp := make([]byte, 0, 32)
	err := DecodeInt(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("TxBlock ID failed %s", err)
	}
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxBlock PrevHash failed %s", err)
	}
	copy(a.PrevHash[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxBlock HashID failed %s", err)
	}
	copy(a.HashID[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxBlock MerkleRoot failed: %s", err)
	}
	copy(a.MerkleRoot[:], tmp[:32])
	err = DecodeInt(buf, &a.Timestamp)
	if err != nil {
		return fmt.Errorf("TxBlock Timestamp failed: %s", err)
	}
	err = DecodeInt(buf, &a.Height)
	if err != nil {
		return fmt.Errorf("TxBlock Height failed: %s", err)
	}
	err = DecodeInt(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxBlock TxCnt failed: %s", err)
	}
	err = DecodeInt(buf, &a.Kind)
	if err != nil {
		return fmt.Errorf("TxBlock Kind failed: %s", err)
	}
	a.TxArray = make([][sHash]byte, 0, a.TxCnt)
	var xxx []byte
	for i := uint32(0); i < a.TxCnt; i++ {
		err = DecodeByteL(buf, &xxx, sHash)
		var tmp [sHash]byte
		copy(tmp[:], xxx[:sHash])
		if err != nil {
			return fmt.Errorf("TxBlock decode Tx failed-%d: %s", i, err)
		}
		a.TxArray = append(a.TxArray, tmp)
	}
	if a.HashID != sha256.Sum256([]byte(GenesisTxBlock)) {
		err = a.Sig.DataToSign(buf)
		if err != nil {
			return fmt.Errorf("TxBlock Signature failed: %s", err)
		}
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxBlock decode failed: With extra bits")
	}
	return nil
}
