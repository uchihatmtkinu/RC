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
	PrevSyncBlockHash [][32]byte
	PrevTxBlockHashes [][32]byte
	PrevRepBlockHash  [32]byte
	Hash              [32]byte
	Nonce             int
}

// NewBlock creates and returns Block
func NewRepBlock(ms *[]shard.MemShard, startBlock bool, prevSyncRepBlockHash [][32]byte, prevTxBlockHashes [][32]byte, prevRepBlockHash [32]byte) (*RepBlock, bool) {
	var item *shard.MemShard
	var repTransactions []*RepTransaction
	tmpprevSyncRepBlockHash:=make([][32]byte,len(prevSyncRepBlockHash))
	copy(tmpprevSyncRepBlockHash,prevSyncRepBlockHash)

	tmpprevTxBlockHashes:=make([][32]byte,len(prevTxBlockHashes))
	copy(tmpprevTxBlockHashes,prevTxBlockHashes)

	for i := uint32(0); i < gVar.ShardSize; i++ {
		item = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		repTransactions = append(repTransactions, NewRepTransaction(shard.ShardToGlobal[shard.MyMenShard.Shard][i], item.Rep))
	}
	var block *RepBlock
	//generate new block
	if startBlock {
		block = &RepBlock{time.Now().Unix(), repTransactions, startBlock, tmpprevSyncRepBlockHash, tmpprevTxBlockHashes, [32]byte{gVar.MagicNumber}, [32]byte{}, 0}
	} else {
		block = &RepBlock{time.Now().Unix(), repTransactions, startBlock, [][32]byte{{gVar.MagicNumber}}, tmpprevTxBlockHashes, prevRepBlockHash, [32]byte{}, 0}
	}
	pow := NewProofOfWork(block)
	nonce, hash, flag := pow.Run()
	block.Hash = hash
	block.Nonce = nonce

	return block, flag
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisRepBlock() *RepBlock {
	block := &RepBlock{time.Now().Unix(), nil, true, [][32]byte{{gVar.MagicNumber}}, [][32]byte{{gVar.MagicNumber}}, [32]byte{gVar.MagicNumber},[32]byte{gVar.MagicNumber}, int(gVar.MagicNumber)}
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
	fmt.Println("RepBlock:")
	fmt.Println("RepTransactions:")
	for _,item := range b.RepTransactions{
		fmt.Print("	GlobalID:", item.GlobalID)
		fmt.Println("		Rep",item.Rep)
	}
	fmt.Println("StartBlock:", b.StartBlock)
	fmt.Println("PrevTxBlockHashes:", b.PrevTxBlockHashes)
	fmt.Println("PrevSyncRepBlockHash",b.PrevSyncBlockHash)
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
