package Reputation

type RepTransaction struct {
	//ID   	[]byte
	MinerID		int
	Trustiness	byte
}

func NewRepTransaction(minerID int, trustiness byte) *RepTransaction{
	tx := RepTransaction{minerID,trustiness}
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