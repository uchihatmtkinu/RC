package Reputation

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"

	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"

	"fmt"
)

// RepBlock reputation block
type RepBlock struct {
	Timestamp       int64
	RepTransactions []*RepTransaction
	StartBlock		bool
	/*
	PrevSyncRepBlockHash [][]byte
	 */
	PrevTxBlockHashes [][32]byte
	PrevRepBlockHash  [32]byte
	Hash              [32]byte
	Nonce             int
}

// NewBlock creates and returns Block
func NewRepBlock(ms *[]shard.MemShard, startBlock bool, prevTxBlockHashes [][32]byte, prevRepBlockHash [32]byte) *RepBlock {
	var item *shard.MemShard
	var repTransactions []*RepTransaction
	for i := uint32(0); i < gVar.ShardSize; i++ {
		item = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		repTransactions = append(repTransactions, NewRepTransaction(shard.ShardToGlobal[shard.MyMenShard.Shard][i], item.Rep))
	}

	//generate new block
	block := &RepBlock{time.Now().Unix(), repTransactions, startBlock, prevTxBlockHashes, prevRepBlockHash, [32]byte{}, 0}

	pow := NewProofOfWork(block)
	nonce, hash, flag := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	if flag {
		RepPowTxCh <- *block
	}
	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisRepBlock() *RepBlock {
	var tmp [][32]byte
	tmp = append(tmp, [32]byte{0})
	block := &RepBlock{time.Now().Unix(), nil, true, tmp, [32]byte{0}, [32]byte{}, 0}
	block.Hash = [32]byte{66}
	return block
}

// HashRep returns a hash of the RepValue in the block
func (b *RepBlock) HashRep() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _, item := range b.RepTransactions {
		txHashes = append(txHashes, IntToHex(item.Rep)[:]...)
		txHashes = append(txHashes, IntToHex(int64(item.GlobalID))...)
	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}

// HashPrevTxBlockHashes returns a hash of the prevTxBlocks in the block
func (b *RepBlock) HashPrevTxBlockHashes() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _, item := range b.PrevTxBlockHashes {
		txHashes = append(txHashes, item[:]...)
	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}

// Serialize encode block
func (b *RepBlock) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (b *RepBlock) Print() {
	fmt.Println("RepTransactions:")
	for _,item := range b.RepTransactions{
		fmt.Print("	GlobalID:", item.GlobalID)
		fmt.Println("		Rep",item.Rep)
	}
	fmt.Println("StartBlock:", b.StartBlock)
	fmt.Println("PrevTxBlockHashes:", b.PrevTxBlockHashes)
	fmt.Println("PrevRepBlockHash:", b.PrevRepBlockHash)
	fmt.Println("Hash:", b.Hash)

}


// DeserializeRepBlock decode Repblock
func DeserializeRepBlock(d []byte) *RepBlock {
	var block RepBlock
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
