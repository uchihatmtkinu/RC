package newrep

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/uchihatmtkinu/RC/basic"
)

//Make create a new RepMsg type
func (a *RepMsg) Make(id uint32, tb [][32]byte, vote []NewRep, round uint32, prikey *ecdsa.PrivateKey) {
	a.ID = id
	a.TBHash = tb
	a.Vote = vote
	a.Round = round
	a.NumTB = uint32(len(a.TBHash))
	a.NumVote = uint32(len(a.Vote))
	tmp := a.Hash()
	a.Sig.Sign(tmp[:], prikey)
}

//Hash returns the hash of the RepMsg
func (a *RepMsg) Hash() [32]byte {
	var tmp []byte
	a.Encode(&tmp)
	tmpHash := new([32]byte)
	basic.DoubleHash256(&tmp, tmpHash)
	return *tmpHash
}

//Encode encode the RepMsg into []byte
func (a *RepMsg) Encode(tmp *[]byte) {
	basic.Encode(tmp, a.ID)
	basic.Encode(tmp, a.Round)
	basic.Encode(tmp, a.NumTB)
	basic.Encode(tmp, a.NumVote)
	for i := uint32(0); i < a.NumTB; i++ {
		basic.Encode(tmp, &a.TBHash[i])
	}
	for i := uint32(0); i < a.NumVote; i++ {
		a.Vote[i].Encode(tmp)
	}
	basic.Encode(tmp, &a.Sig)
}

//Decode decode the []byte into RepMsg
func (a *RepMsg) Decode(buf *[]byte) error {
	err := basic.Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("RepMsg ID decode failed: %s", err)
	}
	err = basic.Decode(buf, &a.Round)
	if err != nil {
		return fmt.Errorf("RepMsg Round decode failed: %s", err)
	}
	err = basic.Decode(buf, &a.NumTB)
	if err != nil {
		return fmt.Errorf("RepMsg NumTB decode failed: %s", err)
	}
	err = basic.Decode(buf, &a.NumVote)
	if err != nil {
		return fmt.Errorf("RepMsg NumVote decode failed: %s", err)
	}
	a.TBHash = make([][32]byte, a.NumTB)
	for i := uint32(0); i < a.NumTB; i++ {
		err = basic.Decode(buf, &a.TBHash[i])
		if err != nil {
			return fmt.Errorf("RepMsg TBHash decode failed: %s", err)
		}
	}
	a.Vote = make([]NewRep, a.NumVote)
	for i := uint32(0); i < a.NumVote; i++ {
		err = a.Vote[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("RepMsg Vote decode failed: %s", err)
		}
	}
	err = basic.Decode(buf, &a.Sig)
	if err != nil {
		return fmt.Errorf("RepMsg Sig decode failed: %s", err)
	}
	return nil
}

//Encode encode the NewRep into []byte
func (v *NewRep) Encode(tmp *[]byte) {
	basic.Encode(tmp, v.ID)
	basic.Encode(tmp, v.Rep)
}

//Decode decode the []byte into NewRep
func (v *NewRep) Decode(buf *[]byte) error {
	err := basic.Decode(buf, &v.ID)
	if err != nil {
		return fmt.Errorf("NewRep ID decode failed: %s", err)
	}
	err = basic.Decode(buf, &v.Rep)
	if err != nil {
		return fmt.Errorf("NewRep Rep decode failed: %s", err)
	}
	return nil
}

//Encode encode the GossipFirMsg into []byte
func (g *GossipFirMsg) Encode(tmp *[]byte) {
	basic.Encode(tmp, g.Cnt)
	for i := uint32(0); i < g.Cnt; i++ {
		g.Data[i].Encode(tmp)
	}
}

//Decode decode the []byte into GossipFirMsg
func (g *GossipFirMsg) Decode(buf *[]byte) error {
	err := basic.Decode(buf, &g.Cnt)
	if err != nil {
		return fmt.Errorf("GossipFirMsg Cnt decode failed: %s", err)
	}
	for i := uint32(0); i < g.Cnt; i++ {
		err = g.Data[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("GossipFirMsg Data decode failed: %s", err)
		}
	}
	if len(*buf) != 0 {
		return fmt.Errorf("GossipFirMsg decode failed: With extra bits")
	}
	return nil
}

//Encode encode the RepSecMsg into []byte
func (a *RepSecMsg) Encode(tmp *[]byte) {
	basic.Encode(tmp, a.ID)
	basic.Encode(tmp, a.Round)
	basic.Encode(tmp, a.NumData)
	for i := uint32(0); i < a.NumData; i++ {
		basic.Encode(tmp, &a.MsgHash[i])
	}
	for i := uint32(0); i < a.NumData; i++ {
		basic.Encode(tmp, &a.MsgSig[i])
	}
	basic.Encode(tmp, &a.Sig)
}

//Decode decode the []byte into RepSecMsg
func (a *RepSecMsg) Decode(buf *[]byte) error {
	err := basic.Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("RepSecMsg ID decode failed: %s", err)
	}
	err = basic.Decode(buf, &a.Round)
	if err != nil {
		return fmt.Errorf("RepSecMsg Round decode failed: %s", err)
	}
	err = basic.Decode(buf, &a.NumData)
	if err != nil {
		return fmt.Errorf("RepSecMsg NumData decode failed: %s", err)
	}
	a.MsgHash = make([][32]byte, a.NumData)
	for i := uint32(0); i < a.NumData; i++ {
		err = basic.Decode(buf, &a.MsgHash[i])
		if err != nil {
			return fmt.Errorf("RepSecMsg MsgHash decode failed: %s", err)
		}
	}
	a.MsgSig = make([]basic.RCSign, a.NumData)
	for i := uint32(0); i < a.NumData; i++ {
		err = basic.Decode(buf, &a.MsgSig[i])
		if err != nil {
			return fmt.Errorf("RepSecMsg MsgSig decode failed: %s", err)
		}
	}
	err = basic.Decode(buf, &a.Sig)
	if err != nil {
		return fmt.Errorf("RepSecMsg Sig decode failed: %s", err)
	}
	return nil
}

//Encode encode the GossipSecMsg into []byte
func (g *GossipSecMsg) Encode(tmp *[]byte) {
	basic.Encode(tmp, g.Cnt)
	for i := uint32(0); i < g.Cnt; i++ {
		g.Data[i].Encode(tmp)
	}
}

//Decode decode the []byte into GossipSecMsg
func (g *GossipSecMsg) Decode(buf *[]byte) error {
	err := basic.Decode(buf, &g.Cnt)
	if err != nil {
		return fmt.Errorf("GossipSecMsg Cnt Read failed: %s", err)
	}
	for i := uint32(0); i < g.Cnt; i++ {
		err = g.Data[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("GossipSecMsg Data Read failed: %s", err)
		}
	}
	if len(*buf) != 0 {
		return fmt.Errorf("GossipSecMsg decode failed: With extra bits")
	}
	return nil
}
