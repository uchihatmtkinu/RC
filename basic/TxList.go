package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
)

//Hash returns the ID of the TxList
func (a *TxList) Hash() [32]byte {
	tmp := make([]byte, 0, a.TxCnt*32)
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp = append(tmp, a.TxArray[i][:]...)
	}
	return sha256.Sum256(tmp)
}

//Sign signs the TxList with the leader's private key
func (a *TxList) Sign(prk *ecdsa.PrivateKey) {
	a.HashID = a.Hash()
	a.Sig.Sign(a.HashID[:], prk)
}

//Verify verify the signature
func (a *TxList) Verify(puk *ecdsa.PublicKey) bool {
	if a.Hash() != a.HashID {
		return false
	}
	return a.Sig.Verify(a.HashID[:], puk)
}

//Set init an instance of TxList given those parameters
func (a *TxList) Set(ID uint32) {
	a.ID = ID
	a.TxCnt = 0
	a.TxArray = nil
}

//AddTx adds the tx into transaction list
func (a *TxList) AddTx(tx *Transaction) {
	a.TxCnt++
	a.TxArray = append(a.TxArray, tx.Hash)
}

//Encode returns the byte of a TxList
func (a *TxList) Encode(tmp *[]byte) {
	EncodeInt(tmp, a.ID)
	EncodeByteL(tmp, a.HashID[:], 32)
	EncodeInt(tmp, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		EncodeByteL(tmp, a.TxArray[i][:], 32)
	}
	a.Sig.SignToData(tmp)
}

//Serial outputs a serial of []byte
func (a *TxList) Serial() []byte {
	var tmp []byte
	a.Encode(&tmp)
	return tmp
}

//Decode decodes the TxList with []byte
func (a *TxList) Decode(buf *[]byte) error {
	tmp := make([]byte, 0, 32)
	err := DecodeInt(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("TxList ID decode failed: %s", err)
	}
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxList HashID decode failed: %s", err)
	}
	copy(a.HashID[:], tmp[:32])
	err = DecodeInt(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxList TxCnt decode failed: %s", err)
	}
	a.TxArray = make([][32]byte, 0, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		err = DecodeByteL(buf, &tmp, 32)
		var xxx [32]byte
		copy(xxx[:], tmp[:32])
		if err != nil {
			return fmt.Errorf("TxList Tx decode failed-%d: %s", i, err)
		}
		a.TxArray = append(a.TxArray, xxx)
	}
	err = a.Sig.DataToSign(buf)
	if err != nil {
		return fmt.Errorf("TxList signature decode failed: %s", err)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxList decode failed: With extra bits")
	}
	return nil
}
