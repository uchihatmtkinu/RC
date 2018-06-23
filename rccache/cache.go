package rccache

import (
	"bytes"
	"crypto/ecdsa"
	"strconv"
	"strings"

	"github.com/uchihatmtkinu/RC/gVar"

	"github.com/uchihatmtkinu/RC/basic"
)

const dbFilex = "TxBlockchain.db"

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
	ID        uint32
	DB        TxBlockChain
	TXCache   map[[32]byte]*CrossShardDec
	HashCache map[[sHash]byte][][32]byte
	LastShard uint32
	ShardNum  uint32

	//Leader
	//TLCache  []basic.TxList
	TLSCache [][gVar.ShardCnt]basic.TxList
	TDSCache [][gVar.ShardCnt]basic.TxDecSet
	TLIndex  map[[32]byte]uint32
	TLS      *[gVar.ShardCnt]basic.TxList
	TDS      *[gVar.ShardCnt]basic.TxDecSet
	//TL       *basic.TxList
	Ready []basic.Transaction
	TxB   *basic.TxBlock
	prk   ecdsa.PrivateKey

	//Miner
	TLNow      *basic.TxDecision
	TLSent     *basic.TxDecision
	startIndex int
	lastIndex  int

	TBCache *[][32]byte

	Leader uint32
}

//New is the initilization of DbRef
func (d *DbRef) New(x uint32, prk ecdsa.PrivateKey) {
	d.ID = x
	d.prk = prk
	d.DB.NewBlockchain(strings.Join([]string{strconv.Itoa(int(d.ID)), dbFilex}, ""))
	d.TXCache = make(map[[6]byte]*CrossShardDec)
	d.TxB = d.DB.LatestTxBlock()
	//d.TL = nil
	//d.TLCache = nil
	d.TLS = nil
	d.TLSCache = nil
	d.TDSCache = nil
	d.startIndex = 0
	d.lastIndex = -1
}

//CrossShardDec  is the database of cache
type CrossShardDec struct {
	Data     *basic.Transaction
	Decision [gVar.ShardSize]byte
	InCheck  []int //-1: Output related
	//0: unknown, Not-related; 1: Yes; 2: No; 3: Related-noresult
	ShardRelated []uint32
	Res          int8 //0: unknown; 1: Yes; -1: No; -2: Can be deleted
	InCheckSum   int
	Total        int
	Value        uint32
}
