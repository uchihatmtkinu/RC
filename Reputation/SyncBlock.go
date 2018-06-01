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
	//Userlist			[]int
	PrevRepBlockHash 	[]byte
	PrevTxBlockHashList	[][]byte
	CoSignature			[]byte
	Hash          	 	[]byte
}

//TODO - add userlist
func NewSynBlock(Userlist []int, PrevRepBlockHash []byte, PrevTxBlockHashList [][]byte) *SyncBlock{
	block := &SyncBlock{time.Now().Unix(), PrevRepBlockHash, PrevTxBlockHashList,nil, []byte{}}
	blockhash := sha256.Sum256(block.prepareData())
	block.Hash = blockhash[:]
	return block
}

func (b *SyncBlock) prepareData() []byte {
	data := bytes.Join(
		[][]byte{
			//IntToHex(b.Userlist),
			b.PrevRepBlockHash,
			b.hashPrevTxBlockHashList(),
			b.CoSignature,
			IntToHex(b.Timestamp),
		},
		[]byte{},
	)

	return data
}

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
}

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


