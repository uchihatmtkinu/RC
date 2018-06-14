package Reputation

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"github.com/uchihatmtkinu/RC/shard"
)

type RepBlock struct {
	Timestamp     	 int64
	RepTransactions	 []*RepTransaction
	/* TODO
	PrevSyncRepBlockHash [][]byte
	PrevTxBlockHashes	[][]byte
	TODO END*/
	PrevRepBlockHash []byte
	Hash          	 []byte
	Nonce         	 int
}

// NewBlock creates and returns Block
func NewRepBlock(ms *[]shard.MemShard, prevRepBlockHash []byte) *RepBlock {

	var repTransactions []*RepTransaction
	for _,item := range *ms{
		repTransactions = append(repTransactions,NewRepTransaction(item.RealAccount.AddrReal,item.Rep))
	}

	//generate new block
	block := &RepBlock{time.Now().Unix(), repTransactions, prevRepBlockHash, []byte{}, 0}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisRepBlock() *RepBlock {
	return NewRepBlock(nil,[]byte{})
}


// HashTransactions returns a hash of the transactions in the block
func (b *RepBlock) HashRep() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _,item := range b.RepTransactions {
		txHashes = append(txHashes, UIntToHex(item.Rep)[:]...)
		txHashes = append(txHashes, item.AddrReal[:]...)
	}
	txHash = sha256.Sum256(txHashes)
	return txHash[:]
}

//encode block
func (b *RepBlock) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

//decode Repblock
func DeserializeRepBlock(d []byte) *RepBlock {
	var block RepBlock
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}