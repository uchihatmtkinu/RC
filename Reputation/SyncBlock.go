package Reputation

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/ed25519"
)

//SyncBlock syncblock
type SyncBlock struct {
	Timestamp     	 	int64
	PrevRepBlockHash 	[32]byte
	PrevSyncBlockHash 	[][32]byte
	IDlist				[]int
	TotalRep			[][] int64//*RepTransaction
	CoSignature			[]byte
	Hash          	 	[32]byte
}

// NewSynBlock new sync block
func NewSynBlock(ms *[]shard.MemShard, prevSyncBlockHash[][32]byte, prevRepBlockHash [32]byte, coSignature []byte) *SyncBlock{
	var item *shard.MemShard
	var repList [][]int64
	var idList	[]int
	//repList = make([][gVar.SlidingWindows]int64, 0)
	for i := 0; i < int(gVar.ShardSize); i++{
		item = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		idList = append(idList,shard.ShardToGlobal[shard.MyMenShard.Shard][i])
		repList = append(repList, item.TotalRep)
	}

	block := &SyncBlock{time.Now().Unix(), prevRepBlockHash,prevSyncBlockHash, idList, repList,coSignature, [32]byte{}}
	block.Hash = sha256.Sum256(block.prepareData())
	return block
}

// prepareData prepare []byte data
func (b *SyncBlock) prepareData() []byte {
	data := bytes.Join(
		[][]byte{
			b.PrevRepBlockHash[:],
			b.HashPrevSyncBlock(),
			b.HashIDList(),
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
	for i := range b.TotalRep {
		for _,item := range b.TotalRep[i] {
			txHashes = append(txHashes, IntToHex(item)[:]...)
		}

	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}


// HashIDList returns a hash of the IDList in the block
func (b *SyncBlock) HashIDList() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _,item := range b.IDlist {
		txHashes = append(txHashes, IntToHex(int64(item))[:]...)
	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}

// HashPrevSyncBlock returns a hash of the previous sync block hash
func (b *SyncBlock) HashPrevSyncBlock() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _,item := range b.PrevSyncBlockHash {
		txHashes = append(txHashes, item[:]...)
	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}


// VerifyCosign verify CoSignature
func (b *SyncBlock) VerifyCoSignature(ms *[]shard.MemShard) bool {
	var	pubKeys		[]ed25519.PublicKey
	var it *shard.MemShard
	pubKeys = make([]ed25519.PublicKey, shard.NumMems)
	for i:=0; i < shard.NumMems; i++ {
		it = &(*ms)[b.IDlist[i]]
		pubKeys[it.InShardId] = it.CosiPub
	}
	sbMessage := b.PrevRepBlockHash[:]
	valid := cosi.Verify(pubKeys, nil, sbMessage, b.CoSignature)
}

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


