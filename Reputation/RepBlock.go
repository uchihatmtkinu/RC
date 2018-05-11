package Reputation

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
)

type RepBlock struct {
	Timestamp     int64
	Data          []byte
	/* TODO
	RepMatrix	  [][]byte
	PrevSyncRepBlockHash [][]byte
	TODO END*/
	PrevRepBlockHash []byte
	Hash          []byte
	Nonce         int
}

// NewBlock creates and returns Block
func NewRepBlock(data string, PrevRepBlockHash []byte) *RepBlock {
	block := &RepBlock{time.Now().Unix(), []byte(data), PrevRepBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisRepBlock() *RepBlock {
	return NewRepBlock("Genesis Block", []byte{})
}

//RepMatrix TODO

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

//decode block
func DeserializeBlock(d []byte) *RepBlock {
	var block RepBlock
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}