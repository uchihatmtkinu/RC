package basic

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

//Set initiates the TxDecision given the TxList and the account
func (a *TxDecision) Set(ID [32]byte, b *TxList) error {
	a.TxCnt = 0
	a.HashID = b.HashID
	a.ID = ID
	return nil
}

//Add adds one decision given the result
func (a *TxDecision) Add(x bool) error {
	tmpNum := a.TxCnt % 8
	a.TxCnt++
	if tmpNum == 0 {
		a.Decision = append(a.Decision, byte(0))
	}
	tmp := len(a.Decision)
	if x {
		a.Decision[tmp] += 1 << tmpNum
	}
	return nil
}

//Sign signs the TxDecision
func (a *TxDecision) Sign(prk *ecdsa.PrivateKey) {
	var tmp []byte
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, a.ID[:]...)
	a.SignR = new(big.Int)
	a.SignS = new(big.Int)
	tmpHash := sha256.Sum256(tmp)
	a.SignR, a.SignS, _ = ecdsa.Sign(rand.Reader, prk, tmpHash[:])
}

//Verify the signature using public key
func (a *TxDecision) Verify(puk *ecdsa.PublicKey) bool {
	var tmp []byte
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, a.ID[:]...)
	tmpHash := sha256.Sum256(tmp)
	return ecdsa.Verify(puk, tmpHash[:], a.SignR, a.SignS)
}

//TDToData encodes the TxDecision into []byte
func (a *TxDecision) TDToData() []byte {
	var tmp []byte
	EncodeByteL(&tmp, a.ID[:], 32)
	EncodeByteL(&tmp, a.HashID[:], 32)
	EncodeInt(&tmp, a.TxCnt)
	EncodeByte(&tmp, &a.Decision)
	EncodeDoubleBig(&tmp, a.SignR, a.SignS)
	return tmp
}

//DataToTD decodes the []byte into TxDecision
func (a *TxDecision) DataToTD(buf *[]byte) error {
	var tmp []byte
	err := DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxDecision ID decode failed %s", err)
	}
	copy(a.ID[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxDecision HashID decode failed %s", err)
	}
	copy(a.HashID[:], tmp[:32])
	err = DecodeInt(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxDecision TxCnt decode failed %s", err)
	}
	err = DecodeByte(buf, &a.Decision)
	if err != nil {
		return fmt.Errorf("TxDecision Decision decode failed %s", err)
	}
	a.SignR = new(big.Int)
	a.SignS = new(big.Int)
	err = DecodeDoubleBig(buf, a.SignR, a.SignS)
	if err != nil {
		return fmt.Errorf("TxDecision Signature decode failed %s", err)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxDecision decode failed: With extra bits %s", err)
	}
	return nil
}
