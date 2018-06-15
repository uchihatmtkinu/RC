package rccache

import (
	"bytes"
	"crypto/ecdsa"

	"github.com/uchihatmtkinu/RC/shard"

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

//DbRef is the structure stores the cache of a miner for the database
type DbRef struct {
	ID       uint32
	db       TxBlockChain
	TXCache  map[[32]byte]*CrossShardDec
	ShardNum uint32

	//Leader
	TLCache  []basic.TxList
	TLSCache [][basic.ShardCnt]basic.TxList
	TDSCache [][basic.ShardCnt]basic.TxDecSet
	TLIndex  map[[32]byte]uint32
	TLS      *[basic.ShardCnt]basic.TxList
	TDS      *[basic.ShardCnt]basic.TxDecSet
	TL       *basic.TxList
	Ready    []basic.Transaction
	TxB      *basic.TxBlock
	prk      ecdsa.PrivateKey

	//Miner
	TLNow      *basic.TxDecision
	TLSent     *basic.TxDecision
	startIndex int
	lastIndex  int

	TBCache *[][32]byte
	Mem     *[]shard.MemShard
}

//New is the initilization of DbRef
func (d *DbRef) New() {
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
	Data     *basic.Transaction
	Decision [basic.ShardSize]byte
	InCheck  []int //-1: Output related
	//0: unknown, Not-related; 1: Yes; 2: No; 3: Related-noresult
	ShardRelated []uint32
	Res          int8 //0: unknown; 1: Yes; 2: No
	InCheckSum   int
	Total        int
	Value        uint32
}
