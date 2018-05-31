package rccache

import (
	"bytes"
	"crypto/ecdsa"

	"github.com/uchihatmtkinu/RC/basic"
)

const dbFile = "TxBlockchain.db"

//UTXOBucket is the bucket of UTXO
const UTXOBucket = "UTXO"

//ACCBucket is the bucket of UTXO
const ACCBucket = "ACC"

//TBBucket is the bucket of ACC
const TBBucket = "TxBlock"

//byteCompare is the func used for string compare
func byteCompare(a, b interface{}) int {
	switch a.(type) {
	case *basic.AccCache:
		tmp1 := a.(*basic.AccCache).ID
		tmp2 := b.(*basic.AccCache).ID
		return bytes.Compare(tmp1[:], tmp2[:])
	default:
		return bytes.Compare([]byte(a.([]byte)), []byte(b.([]byte)))
	}

}

//dbRef is the structure stores the cache of a miner for the database
type dbRef struct {
	ID       [32]byte
	db       TxBlockChain
	TX       map[[32]byte]*CrossShardDec
	ShardNum uint32
	TL       basic.TxList
	TLS      []basic.TxList
	TDS      []basic.TxDecSet
	Ready    []basic.Transaction
	TxB      *basic.TxBlock
	prk      ecdsa.PrivateKey
}

//New is the initilization of dbRef
func (d *dbRef) New() {
	d.db.NewBlockchain()
	d.TX = make(map[[32]byte]*CrossShardDec)
	d.TxB = d.db.LatestTxBlock()
}

//CrossShardDec  is the database of cache
type CrossShardDec struct {
	Data       basic.Transaction
	InCheck    []bool
	Res        int8
	InCheckSum int
	Yes        uint32
	No         uint32
}
