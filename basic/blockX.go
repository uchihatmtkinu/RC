package basic

import (
	"crypto/sha256"
	"fmt"
)

//Transform change the txlist to txlistX
func (a *TxBlock) Transform() *TxBlockX {
	tmp := new(TxBlockX)
	tmp.ID = a.ID
	tmp.PrevHash = a.PrevHash
	tmp.MerkleRoot = a.MerkleRoot
	tmp.HashID = a.HashID
	tmp.Kind = a.Kind
	tmp.Timestamp = a.Timestamp
	tmp.Height = a.Height
	tmp.TxCnt = a.TxCnt
	tmp.TxArray = make([][sHash]byte, a.TxCnt)
	for i := uint32(0); i < tmp.TxCnt; i++ {
		tmp.TxArray[i] = HashCut(a.TxArray[i].Hash)
	}
	tmp.Sig.New(&a.Sig)
	return tmp
}

//Encode converts the block data into bytes
func (a *TxBlockX) Encode(tmp *[]byte) {
	Encode(tmp, a.ID)
	Encode(tmp, &a.PrevHash)
	Encode(tmp, &a.HashID)
	Encode(tmp, &a.MerkleRoot)
	Encode(tmp, a.Timestamp)
	Encode(tmp, a.Height)
	Encode(tmp, a.TxCnt)
	Encode(tmp, a.Kind)
	for i := uint32(0); i < a.TxCnt; i++ {
		Encode(tmp, a.TxArray[i])
	}
	Encode(tmp, &a.Sig)
}

//Decode converts bytes into block data
func (a *TxBlockX) Decode(buf *[]byte) error {
	err := Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("TxBlock ID failed %s", err)
	}
	err = Decode(buf, &a.PrevHash)
	if err != nil {
		return fmt.Errorf("TxBlock PrevHash failed %s", err)
	}
	err = Decode(buf, &a.HashID)
	if err != nil {
		return fmt.Errorf("TxBlock HashID failed %s", err)
	}
	err = Decode(buf, &a.MerkleRoot)
	if err != nil {
		return fmt.Errorf("TxBlock MerkleRoot failed: %s", err)
	}
	err = Decode(buf, &a.Timestamp)
	if err != nil {
		return fmt.Errorf("TxBlock Timestamp failed: %s", err)
	}
	err = Decode(buf, &a.Height)
	if err != nil {
		return fmt.Errorf("TxBlock Height failed: %s", err)
	}
	err = Decode(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxBlock TxCnt failed: %s", err)
	}
	err = Decode(buf, &a.Kind)
	if err != nil {
		return fmt.Errorf("TxBlock Kind failed: %s", err)
	}
	a.TxArray = make([][sHash]byte, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		err = Decode(buf, &a.TxArray[i])
		if err != nil {
			return fmt.Errorf("TxBlock decode Tx failed-%d: %s", i, err)
		}
	}
	if a.HashID != sha256.Sum256([]byte(GenesisTxBlock)) {
		err = Decode(buf, &a.Sig)
		if err != nil {
			return fmt.Errorf("TxBlock Signature failed: %s", err)
		}
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxBlock decode failed: With extra bits")
	}
	return nil
}
