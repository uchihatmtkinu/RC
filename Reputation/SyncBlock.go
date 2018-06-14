package Reputation

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
)


type SyncBlock struct {
	Timestamp     	 	int64
	Userlist			[][32]byte
	PrevRepBlockHash 	[]byte
	CoSignature			[]byte
	Hash          	 	[]byte
}

//TODO - add userlist
func NewSynBlock(userlist [][32]byte, prevRepBlockHash []byte, coSignature []byte) *SyncBlock{
	block := &SyncBlock{time.Now().Unix(), userlist, prevRepBlockHash,coSignature, []byte{}}
	blockhash := sha256.Sum256(block.prepareData())
	block.Hash = blockhash[:]
	return block
}

func (b *SyncBlock) prepareData() []byte {
	data := bytes.Join(
		[][]byte{
			b.UserlistHash(),
			b.PrevRepBlockHash,
			b.PrevTxBlockHash,
			b.CoSignature,
			//IntToHex(b.Timestamp),
		},
		[]byte{},
	)

	return data
}

// returns a hash of the userlist in the block
func (b *SyncBlock) UserlistHash() []byte {
	var txHashes []byte
	var txHash [32]byte
	for _,item := range b.Userlist {
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

//encode block
func (b *SyncBlock) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}


//decode Syncblock
func DeserializeSyncBlock(d []byte) *SyncBlock {
	var block SyncBlock
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}


