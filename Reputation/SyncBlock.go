package Reputation

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//SyncBlock syncblock
type SyncBlock struct {
	Timestamp     	 	int64
	//userlist			[]int
	PrevRepBlockHash 	[32]byte
	PrevSyncBlockHash 	[][32]byte
	TotalRepTrans		[gVar.SlidingWindows][] *RepTransaction
	CoSignature			[]byte
	Hash          	 	[32]byte
}

// NewSynBlock new sync block
func NewSynBlock(ms *[]shard.MemShard, prevSyncBlockHash[][32]byte, prevRepBlockHash [32]byte, coSignature []byte) *SyncBlock{
	var item *shard.MemShard
	var repTransactions []*RepTransaction
	for k := 0; k < int(gVar.SlidingWindows); k++ {
		for i := 0; i < int(gVar.ShardSize); i++ {
			item = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
			repTransactions = append(repTransactions, NewRepTransaction(shard.ShardToGlobal[shard.MyMenShard.Shard][i], item.TotalRep[k]))
		}
	}

	block := &SyncBlock{time.Now().Unix(), prevRepBlockHash,prevSyncBlockHash, repTransactions,coSignature, [32]byte{}}
	blockhash := sha256.Sum256(block.prepareData())
	block.Hash = blockhash
	return block
}

// prepareData prepare []byte data
func (b *SyncBlock) prepareData() []byte {
	data := bytes.Join(
		[][]byte{
			b.HashPrevSyncBlock(),
			b.PrevRepBlockHash[:],
			b.HashTotalRep(),
			b.CoSignature,
			//IntToHex(b.Timestamp),
		},
		[]byte{},
	)

	return data
}


// HashRep returns a hash of the TotalRepTransactions in the block
func (b *SyncBlock) HashTotalRep() []byte {
	var txHashes []byte
	var txHash [32]byte
	for i := range b.TotalRepTrans {
		for _,item := range b.TotalRepTrans[i] {
			txHashes = append(txHashes, IntToHex(item.Rep)[:]...)
			txHashes = append(txHashes, IntToHex(int64(item.GlobalID))...)
		}

	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}

// UserlistHash returns a hash of the userlist in the block
func (b *SyncBlock) HashPrevSyncBlock() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _,item := range b.PrevSyncBlockHash {
		txHashes = append(txHashes, item[:]...)
	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}

/*
// HashTransactions returns a hash of the transactions in the block
func (b *SyncBlock) hashPrevTxBlockHashList() []byte {
	var blockHashes []byte
	var blockHash [32]byte

	for i := range b.PrevTxBlockHashList {
		for _,value := range b.PrevTxBlockHashList[i]{
			blockHashes = append(blockHashes, value)
		}
	}
	blockHash = sha256.Sum256(bytes.Join([][]byte{blockHashes}, []byte{}))
	return blockHash[:]
}*/

// Serialize encode block
func (b *SyncBlock) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}


// DeserializeSyncBlock decode Syncblock
func DeserializeSyncBlock(d []byte) *SyncBlock {
	var block SyncBlock
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}


