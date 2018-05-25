package basic

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
)

//Build implements the TxDecSS type given the TxDecSet data
func (a *TxDecSS) Build(b *[]TxDecSet) {
	a.ShardNum = uint32(len(*b))
	a.Header = make([]TDSHeader, a.ShardNum)
	tmp := make(map[[32]byte]uint32, 5000)
	a.TxCnt = 0
	a.Tx = make([][32]byte, 0, 5000)
	for i := uint32(0); i < a.ShardNum; i++ {
		a.Header[i].ID = (*b)[i].ID
		a.Header[i].HashID = (*b)[i].HashID
		a.Header[i].PrevHash = (*b)[i].PrevHash
		a.Header[i].MemCnt = (*b)[i].MemCnt
		a.Header[i].Sig = (*b)[i].Sig
		a.Header[i].TxCnt = (*b)[i].TxCnt
		a.Header[i].MemD = make([]TxDPure, a.Header[i].MemCnt)
		a.Header[i].TxIndex = make([]uint32, a.Header[i].TxCnt)
		for j := uint32(0); j < (*b)[i].MemCnt; j++ {
			a.Header[i].MemD[j].ID = (*b)[i].MemD[j].ID
			a.Header[i].MemD[j].Decision = (*b)[i].MemD[j].Decision
			a.Header[i].MemD[j].Sig = (*b)[i].MemD[j].Sig
		}
		for j := uint32(0); j < (*b)[i].TxCnt; j++ {
			tmpValue, ok := tmp[(*b)[i].TxArray[j]]
			if !ok {
				a.Tx = append(a.Tx, (*b)[i].TxArray[j])
				a.TxCnt++
				tmp[(*b)[i].TxArray[j]] = a.TxCnt
				a.Header[i].TxIndex[j] = a.TxCnt - 1
			} else {
				a.Header[i].TxIndex[j] = tmpValue - 1
			}
		}
	}
}

//VerifyHash verfies the hash and data structure of TxDecSS
func (a *TxDecSS) VerifyHash() bool {
	if a.ShardNum != uint32(len(a.Header)) {
		return false
	}
	if a.TxCnt != uint32(len(a.Tx)) {
		return false
	}
	for i := uint32(0); i < a.ShardNum; i++ {
		tmp := make([]byte, 0, a.Header[i].TxCnt*32)
		if a.Header[i].TxCnt != uint32(len(a.Header[i].TxIndex)) {
			return false
		}
		for j := uint32(0); j < a.Header[i].TxCnt; j++ {
			if a.Header[i].TxIndex[j] >= a.TxCnt {
				return false
			}
			tmp = append(tmp, a.Tx[a.Header[i].TxIndex[j]][:]...)
		}
		tmpHash := sha256.Sum256(tmp)
		if tmpHash != a.Header[i].HashID {
			return false
		}
	}
	return true
}

//VerifyHeader verifies the header signature and address
func (a *TxDecSS) VerifyHeader(x int, puk *ecdsa.PublicKey, ID [32]byte) bool {
	if uint32(x) >= a.ShardNum {
		return false
	}
	if a.Header[x].ID != ID {
		return false
	}
	tmp := make([]byte, 0, 64+len(a.Header[x].MemD[0].Decision)*int(a.Header[x].MemCnt))
	tmp = append(a.Header[x].ID[:], a.Header[x].HashID[:]...)
	for i := uint32(0); i < a.Header[x].MemCnt; i++ {
		tmp = append(tmp, a.Header[x].MemD[i].Decision...)
	}
	tmpHash := sha256.Sum256(tmp)
	return a.Header[x].Sig.Verify(tmpHash[:], puk)
}

//VerifyDec verifies the decision signature and address
func (a *TxDecSS) VerifyDec(x int, y int, puk *ecdsa.PublicKey, ID [32]byte) bool {
	if uint32(x) >= a.ShardNum {
		return false
	}
	if uint32(y) >= a.Header[x].MemCnt {
		return false
	}
	if a.Header[x].MemD[y].ID != ID {
		return false
	}
	tmp := make([]byte, 0, 64+len(a.Header[x].MemD[y].Decision))
	tmp = append(a.Header[x].HashID[:], a.Header[x].MemD[y].Decision...)
	tmp = append(tmp, a.Header[x].MemD[y].ID[:]...)
	tmpHash := sha256.Sum256(tmp)
	return a.Header[x].MemD[y].Sig.Verify(tmpHash[:], puk)
}

//Result returns the y-th Tx result by x-th shard and its ID
func (a *TxDecSS) Result(x int, y int, max int) ([32]byte, bool) {
	if uint32(x) >= a.ShardNum {
		return [32]byte{}, false
	}
	if uint32(y) >= a.Header[x].TxCnt {
		return [32]byte{}, false
	}
	total := 0
	index := y / 8
	shift := byte(y % 8)
	for i := uint32(0); i < a.Header[x].MemCnt; i++ {
		total += int((a.Header[x].MemD[i].Decision[index] >> shift) & 1)
		if total > max/2 {
			return a.Tx[a.Header[x].TxIndex[y]], true
		}
	}
	return a.Tx[a.Header[x].TxIndex[y]], false
}

//Encode encodes the TxDPure into []byte
func (a *TxDPure) Encode(b *[]byte) {
	EncodeByteL(b, a.ID[:], 32)
	EncodeByte(b, &a.Decision)
	a.Sig.SignToData(b)
}

//Decode decodes the []byte into TxDPure
func (a *TxDPure) Decode(buf *[]byte) error {
	var tmp []byte
	err := DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TxDPure ID decode failed: %s", err)
	}
	copy(a.ID[:], tmp[:32])
	err = DecodeByte(buf, &a.Decision)
	if err != nil {
		return fmt.Errorf("TxDPure Decision decode failed: %s", err)
	}
	err = a.Sig.DataToSign(buf)
	if err != nil {
		return fmt.Errorf("TxDPure Sig decode failed: %s", err)
	}
	return nil
}

//Encode encodes the TDSHeader into []byte
func (a *TDSHeader) Encode(b *[]byte) {
	EncodeByteL(b, a.ID[:], 32)
	EncodeByteL(b, a.HashID[:], 32)
	EncodeByteL(b, a.PrevHash[:], 32)
	EncodeInt(b, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		EncodeInt(b, a.TxIndex[i])
	}
	EncodeInt(b, a.MemCnt)
	for i := uint32(0); i < a.MemCnt; i++ {
		a.MemD[i].Encode(b)
	}
	a.Sig.SignToData(b)
}

//Decode decodes the []byte into TDSHeader
func (a *TDSHeader) Decode(buf *[]byte) error {
	var tmp []byte
	err := DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TDSHeader ID decode failed: %s", err)
	}
	copy(a.ID[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TDSHeader HashID decode failed: %s", err)
	}
	copy(a.HashID[:], tmp[:32])
	err = DecodeByteL(buf, &tmp, 32)
	if err != nil {
		return fmt.Errorf("TDSHeader PrevHash decode failed: %s", err)
	}
	copy(a.PrevHash[:], tmp[:32])
	err = DecodeInt(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TDSHeader TxCnt decode failed: %s", err)
	}
	a.TxIndex = make([]uint32, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		err = DecodeInt(buf, &a.TxIndex[i])
		if err != nil {
			return fmt.Errorf("TDSHeader TxArray decode failed: %s", err)
		}
	}
	err = DecodeInt(buf, &a.MemCnt)
	if err != nil {
		return fmt.Errorf("TDSHeader MemCnt decode failed: %s", err)
	}
	a.MemD = make([]TxDPure, a.MemCnt)
	for i := uint32(0); i < a.MemCnt; i++ {
		err = a.MemD[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("TDSHeader MemD decode failed-%d: %s", i, err)
		}
	}
	err = a.Sig.DataToSign(buf)
	if err != nil {
		return fmt.Errorf("TDSHeader Sig decode failed: %s", err)
	}
	return nil
}

//Encode encodes the TDSHeader into []byte
func (a *TxDecSS) Encode(b *[]byte) {
	EncodeInt(b, a.ShardNum)
	for i := uint32(0); i < a.ShardNum; i++ {
		a.Header[i].Encode(b)
	}
	EncodeInt(b, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		EncodeByteL(b, a.Tx[i][:], 32)
	}
}

//Decode decodes []byte into TxDecSS
func (a *TxDecSS) Decode(buf *[]byte) error {
	err := DecodeInt(buf, &a.ShardNum)
	if err != nil {
		return fmt.Errorf("TxDecSS ShardNum decode failed: %s", err)
	}
	for i := uint32(0); i < a.ShardNum; i++ {
		err = a.Header[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("TxDecSS TDSHeader decode failed-%d: %s", i, err)
		}
	}
	err = DecodeInt(buf, &a.TxCnt)
	if err != nil {
		return fmt.Errorf("TxDecSS TxCnt decode failed: %s", err)
	}
	var tmp []byte
	var tmp1 [32]byte
	a.Tx = make([][32]byte, 0, a.TxCnt)
	for i := uint32(0); i < a.TxCnt; i++ {
		err = DecodeByteL(buf, &tmp, 32)
		if err != nil {
			return fmt.Errorf("TxDecSS Tx decode failed-%d: %s", i, err)
		}
		copy(tmp1[:], tmp[:32])
		a.Tx = append(a.Tx, tmp1)
	}
	if len(*buf) != 0 {
		return fmt.Errorf("TxDecSS decode failed: With extra bits")
	}
	return nil
}
