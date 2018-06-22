package Reputation

import "github.com/dedis/kyber/sign/cosi"

type SyncRepTransaction struct {
	GlobalID   	int
	Mask		cosi.Mask
	Rep			int64
}

//new reputation transaction
func NewSyncRepTransaction(globalID int, mask cosi.Mask, rep int64) *SyncRepTransaction{
	tx := SyncRepTransaction{globalID,mask, rep}
	return &tx
}


// SetID sets ID of a transaction
/*
func (tx *RepTransaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}
*/