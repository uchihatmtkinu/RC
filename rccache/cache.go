package rccache

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/uchihatmtkinu/RC/base58"
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

//FBBucket is the bucket of Final Blocks
const FBBucket = "FBTxBlock"

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
	Mu            sync.RWMutex
	ID            uint32
	DB            TxBlockChain
	TXCache       map[[32]byte]*CrossShardDec
	HashCache     map[[basic.SHash]byte][][32]byte
	WaitHashCache map[[basic.SHash]byte]WaitProcess

	ShardNum uint32

	//Leader
	//TLCache  []basic.TxList
	TLSCache      [][gVar.ShardCnt]basic.TxList
	TLSCacheMiner map[[32]byte]*basic.TxList
	TDSCache      [][gVar.ShardCnt]basic.TxDecSet
	TLIndex       map[[32]byte]uint32
	TLS           *[gVar.ShardCnt]basic.TxList
	TDS           *[gVar.ShardCnt]basic.TxDecSet
	Ready         []basic.Transaction
	TxB           *basic.TxBlock
	FB            [gVar.ShardCnt]*basic.TxBlock
	prk           ecdsa.PrivateKey
	TxCnt         uint32

	//Miner
	TLNow      *basic.TxDecision
	TLSent     *basic.TxDecision
	StartIndex int
	LastIndex  int
	TLRound    uint32
	PrevHeight uint32

	TBCache *[][32]byte

	Leader        uint32
	UnderSharding bool
	StopGetTx     bool
	StartTxDone   bool
}

//PreStat is used in pre-defined request
type PreStat struct {
	Stat  int
	Valid []int
}

//WaitProcess is the current wait process
type WaitProcess struct {
	DataTB  []*basic.TxBlock
	StatTB  []*PreStat
	IDTB    []int
	DataTL  []*basic.TxList
	StatTL  []*PreStat
	IDTL    []int
	DataTDS []*basic.TxDecSet
	StatTDS []*PreStat
	IDTDS   []int
}

//Clear refresh the data for next epoch
func (d *DbRef) Clear() {
	d.TLRound = 0
	d.TXCache = make(map[[32]byte]*CrossShardDec, 1000)
	d.TLS = nil
	d.TxCnt = 0
	d.HashCache = make(map[[basic.SHash]byte][][32]byte, 10000)
	if len(*d.TBCache) != 0 {
		fmt.Println("Miner", d.ID, "Cache clear: TBCache is not empty")
		for i := 0; i < len(*d.TBCache); i++ {
			fmt.Println("TBCache of", d.ID, "-", i, ":", base58.Encode((*d.TBCache)[i][:]))
		}
	}
}

//New is the initilization of DbRef
func (d *DbRef) New(x uint32, prk ecdsa.PrivateKey) {
	d.ID = x
	d.prk = prk
	d.TxCnt = 0
	d.DB.NewBlockchain(strings.Join([]string{strconv.Itoa(int(d.ID)), dbFilex}, ""))
	d.TXCache = make(map[[32]byte]*CrossShardDec, 1000)
	d.TxB = d.DB.LatestTxBlock()
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		d.FB[i] = d.DB.LatestFinalTxBlock(i)
	}
	//d.TL = nil
	//d.TLCache = nil
	//d.TLS = new([gVar.ShardCnt]basic.TxList)
	d.TLS = nil
	d.TLIndex = make(map[[32]byte]uint32, 100)
	d.WaitHashCache = make(map[[basic.SHash]byte]WaitProcess, 1000)
	d.TDS = new([gVar.ShardCnt]basic.TxDecSet)
	d.TLSCache = nil
	d.TLSCacheMiner = make(map[[32]byte]*basic.TxList, 100)
	d.TDSCache = nil
	d.StartIndex = 0
	d.LastIndex = -1
	d.HashCache = make(map[[basic.SHash]byte][][32]byte, 10000)
	d.TBCache = new([][32]byte)
	d.StopGetTx = false
	d.UnderSharding = true
	d.StartTxDone = true
	d.PrevHeight = 0
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

//Print output the crossshard information
func (c *CrossShardDec) Print() {
	fmt.Println("-----------CrossShardDec-------------")
	fmt.Println("Tx Hash: ", base58.Encode(c.Data.Hash[:]), "Value: ", c.Value)
	fmt.Println("ShardRelated: ", c.ShardRelated)
	fmt.Println("IncheckSum: ", c.InCheckSum, " Detail: ", c.InCheck)
	fmt.Println("Decision: ", c.Decision)
}

//ClearCache is to handle the TxCache of hash
func (d *DbRef) ClearCache(HashID [32]byte) error {
	tmp := basic.HashCut(HashID)
	xxx := d.HashCache[tmp]
	if len(xxx) == 0 && xxx[0] == HashID {
		delete(d.HashCache, tmp)
	} else {
		for i := 0; i < len(xxx); i++ {
			if xxx[i] == HashID {
				xxx = append(xxx[:i], xxx[i+1:]...)
				d.HashCache[tmp] = xxx
			}
		}
	}
	delete(d.TXCache, HashID)
	return nil
}

//AddCache is to handle the TxCache of hash
func (d *DbRef) AddCache(HashID [32]byte) error {
	tmp := basic.HashCut(HashID)
	xxx, ok := d.HashCache[tmp]
	if !ok {
		d.HashCache[tmp] = [][32]byte{HashID}
	} else {
		d.HashCache[tmp] = append(xxx, HashID)
	}
	return nil
}
