package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
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
	tmp := make([]byte, 0, 64+len(a.Decision))
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, a.ID[:]...)
	tmpHash := sha256.Sum256(tmp)
	a.Sig.Sign(tmpHash[:], prk)
}

//Verify the signature using public key
func (a *TxDecision) Verify(puk *ecdsa.PublicKey) bool {
	tmp := make([]byte, 0, 64+len(a.Decision))
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, a.ID[:]...)
	tmpHash := sha256.Sum256(tmp)
	return a.Sig.Verify(tmpHash[:], puk)
}

//Encode encodes the TxDecision into []byte
func (a *TxDecision) Encode(tmp *[]byte) {
	EncodeByteL(tmp, a.ID[:], 32)
	EncodeByteL(tmp, a.HashID[:], 32)
	EncodeInt(tmp, a.TxCnt)
	EncodeByte(tmp, &a.Decision)
	a.Sig.SignToData(tmp)
}

// Decode decodes the []byte into TxDecision
func (a *TxDecision) Decode(buf *[]byte) error {
	var tmp []byte
	err := DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxDecsion ID decode failed: %s", err)
	}
	copy(a.ID[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxDecsion HashID decode failed: %s", err)
	}
	copy(a.HashID[:], tmp[:32])
	err = DecodeInt(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxDecsion TxCnt decode failed: %s", err)
	}
	err = DecodeByte(buf, &a.Decision)
	if err != nil {
		return fmt.Errorf("TxDecsion Decision decode failed: %s", err)
	}
	err = a.Sig.DataToSign(buf)
	if err != nil {
		return fmt.Errorf("TxDecsion Signature decode failed: %s", err)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxDecsion decode failed: With extra bits")
	}
	return nil
}
