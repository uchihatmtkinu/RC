package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
)

//Set initiates the TxDecision given the TxList and the account
func (a *TxDecision) Set(ID [32]byte, b *TxList, index uint32) error {
	a.TxCnt = 0
	a.HashID = b.HashID
	a.ID = ID
	a.Target = index
	return nil
}

//Add adds one decision given the result
func (a *TxDecision) Add(x byte) error {
	tmpNum := a.TxCnt % 8
	a.TxCnt++
	if tmpNum == 0 {
		a.Decision = append(a.Decision, byte(0))
	}
	tmp := len(a.Decision)

	a.Decision[tmp] += x << tmpNum

	return nil
}

//Sign signs the TxDecision
func (a *TxDecision) Sign(prk *ecdsa.PrivateKey, x uint32) {
	tmp := make([]byte, 0, 64+len(a.Decision))
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, a.ID[:]...)
	tmpHash := sha256.Sum256(tmp)
	a.Sig[x].Sign(tmpHash[:], prk)
}

//Verify the signature using public key
func (a *TxDecision) Verify(puk *ecdsa.PublicKey, x uint32) bool {
	tmp := make([]byte, 0, 64+len(a.Decision))
	tmp = append(a.HashID[:], a.Decision...)
	tmp = append(tmp, a.ID[:]...)
	tmpHash := sha256.Sum256(tmp)
	return a.Sig[x].Verify(tmpHash[:], puk)
}

//Encode encodes the TxDecision into []byte
func (a *TxDecision) Encode(tmp *[]byte) {
	EncodeByteL(tmp, a.ID[:], 32)
	EncodeByteL(tmp, a.HashID[:], 32)
	EncodeInt(tmp, a.TxCnt)
	EncodeInt(tmp, a.Target)
	EncodeByte(tmp, &a.Decision)
	EncodeInt(tmp, a.Single)
	if a.Single == 1 {
		a.Sig[0].SignToData(tmp)
	} else {
		for i := uint32(0); i < ShardCnt; i++ {
			a.Sig[i].SignToData(tmp)
		}
	}

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
	err = DecodeInt(buf, &a.Target)
	if err != nil {
		return fmt.Errorf("TxDecsion Target decode failed: %s", err)
	}
	err = DecodeByte(buf, &a.Decision)
	if err != nil {
		return fmt.Errorf("TxDecsion Decision decode failed: %s", err)
	}
	err = DecodeInt(buf, &a.Single)
	if err != nil {
		return fmt.Errorf("TxDecision Single decode failed: %s", err)
	}
	if a.Single == 1 {
		a.Sig = make([]RCSign, 1)
		err = a.Sig[0].DataToSign(buf)
		if err != nil {
			return fmt.Errorf("TxDecision Sig decode failed: %s", err)
		}
	} else {
		a.Sig = make([]RCSign, ShardCnt)
		for i := uint32(0); i < ShardCnt; i++ {
			err = a.Sig[i].DataToSign(buf)
			if err != nil {
				return fmt.Errorf("TxDecision Sig decode failed: %s", err)
			}
		}
	}
	return nil
}
