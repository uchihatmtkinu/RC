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
	ID         [32]byte
	db         TxBlockChain
	TXCache    map[[32]byte]*CrossShardDec
	ShardNum   uint32
	TLCache    []basic.TxList
	TLSCache   [][basic.ShardCnt]basic.TxList
	TDSCache   [][basic.ShardCnt]basic.TxDecSet
	TLIndex    map[[32]byte]uint32
	TLS        *[basic.ShardCnt]basic.TxList
	TDS        *[basic.ShardCnt]basic.TxDecSet
	TL         *basic.TxList
	Ready      []basic.Transaction
	TxB        *basic.TxBlock
	prk        ecdsa.PrivateKey
	TDSS       *basic.TxDecSS
	TDSSSent   *basic.TxDecSS
	startIndex int
	lastIndex  int
}

//New is the initilization of dbRef
func (d *dbRef) New() {
	d.db.NewBlockchain()
	d.TXCache = make(map[[32]byte]*CrossShardDec)
	d.TxB = d.db.LatestTxBlock()
	d.TL = nil
	d.TLCache = nil
	d.TLSCache = nil
	d.TDSCache = nil
	d.startIndex = 0
	d.lastIndex = -1
}

//CrossShardDec  is the database of cache
type CrossShardDec struct {
	Data    *basic.Transaction
	InCheck []int //-1: Output related
	//0: unknown, Not-related; 1: Yes; 2: No; 3: Related-noresult
	ShardRelated []uint32
	Res          int8 //0: unknown; 1: Yes; 2: No
	InCheckSum   int
	Yes          uint32
	No           uint32
	Total        int
}
