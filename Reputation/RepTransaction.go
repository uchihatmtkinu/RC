package Reputation

type RepTransaction struct {
	//ID   	[]byte
	AddrReal 	[32]byte //public key -> id
	Rep			uint64
}

//new reputation transaction
func NewRepTransaction(add [32]byte, rep uint64) *RepTransaction{
	tx := RepTransaction{add,rep}
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