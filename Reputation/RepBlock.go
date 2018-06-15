package Reputation

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/testforclient/network"
	"github.com/uchihatmtkinu/RC/gVar"
)

type RepBlock struct {
	Timestamp     	 int64
	RepTransactions	 []*RepTransaction
	/* TODO
	PrevSyncRepBlockHash [][]byte
	TODO END*/
	PrevTxBlockHashes	[][32]byte
	PrevRepBlockHash []byte
	Hash          	 []byte
	Nonce         	 int
}

// NewBlock creates and returns Block
func NewRepBlock(ms *[]shard.MemShard, prevTxBlockHashes *[][32]byte, prevRepBlockHash []byte) *RepBlock {
	var item *shard.MemShard
	var repTransactions []*RepTransaction
	for i:=uint32(0); i <gVar.ShardSize; i++ {
		item = &(*ms)[network.ShardToGlobal[network.MyMenShard.Shard][i]]
		repTransactions = append(repTransactions,NewRepTransaction(item.RealAccount.AddrReal,item.Rep))
	}

	//generate new block
	block := &RepBlock{time.Now().Unix(), repTransactions, *prevTxBlockHashes, prevRepBlockHash, []byte{}, 0}

	pow := NewProofOfWork(block)
	nonce, hash, flag := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	if flag {
		RepPowTxCh <- *block
	}
	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisRepBlock() *RepBlock {
	return NewRepBlock(nil,nil,[]byte{0})
}


// returns a hash of the RepValue in the block
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

// returns a hash of the prevTxBlocks in the block
func (b *RepBlock) HashPrevTxBlockHashes() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _,item := range b.PrevTxBlockHashes {
		txHashes = append(txHashes, item[:]...)
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