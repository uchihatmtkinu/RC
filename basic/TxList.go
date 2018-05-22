package basic

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

//Hash returns the ID of the TxList
func (a *TxList) Hash() [32]byte {
	tmp := a.TxArray[0].Hash[:]
	for i := uint32(1); i < a.TxCnt; i++ {
		tmp = append(tmp, a.TxArray[i].Hash[:]...)
	}
	return sha256.Sum256(tmp)
}

//Sign signs the TxList with the leader's private key
func (a *TxList) Sign(prk *ecdsa.PrivateKey) {
	a.SignR = new(big.Int)
	a.SignS = new(big.Int)
	a.SignR, a.SignS, _ = ecdsa.Sign(rand.Reader, prk, a.HashID[:])
}

//Verify verify the signature
func (a *TxList) Verify(puk *ecdsa.PublicKey) bool {
	return ecdsa.Verify(puk, a.HashID[:], a.SignR, a.SignS)
}

//Set init an instance of TxList given those parameters
func (a *TxList) Set(ID [32]byte, prevH [32]byte) {
	a.ID = ID
	a.PrevHash = prevH
	a.TxCnt = 0
	a.TxArray = nil
}

//AddTx adds the tx into transaction list
func (a *TxList) AddTx(tx *Transaction) {
	a.TxCnt++
	a.TxArray = append(a.TxArray, *tx)
}

//TLToData returns the byte of a TxList
func (a *TxList) TLToData() []byte {
	var tmp []byte
	EncodeByteL(&tmp, a.ID[:], 32)
	EncodeByteL(&tmp, a.HashID[:], 32)
	EncodeByteL(&tmp, a.PrevHash[:], 32)
	EncodeInt(&tmp, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp = append(tmp, a.TxArray[i].TxToData()...)
	}
	EncodeDoubleBig(&tmp, a.SignR, a.SignS)
	return tmp
}

//DataToTL decodes the TxList with []byte
func (a *TxList) DataToTL(buf *[]byte) error {
	var tmp []byte
	err := DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxList ID decode failed %s", err)
	}
	copy(a.ID[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxList HashID decode failed %s", err)
	}
	copy(a.HashID[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxList PrevHash decode failed %s", err)
	}
	copy(a.PrevHash[:], tmp[:32])
	err = DecodeInt(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxList TxCnt decode failed %s", err)
	}
	for i := uint32(0); i < a.TxCnt; i++ {
		var xxx Transaction
		err = xxx.DataToTx(buf)
		if err != nil {
			return fmt.Errorf("TxList Tx decode failed %s", err)
		}
		a.AddTx(&xxx)
	}
	err = DecodeDoubleBig(buf, a.SignR, a.SignS)
	if err != nil {
		return fmt.Errorf("TxList signature decode failed %s", err)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxList decode failed: With extra bits %s", err)
	}
	return nil
}
