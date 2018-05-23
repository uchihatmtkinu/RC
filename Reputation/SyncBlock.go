package Reputation

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
)


type SyncBlock struct {
	Timestamp     	 	int64
	PrevRepBlockHash 	[]byte
	PrevTxBlockHashList	[][]byte
	CoSignature			[]byte
	Hash          	 	[]byte
}


func NewSynBlock(Userlist []int, PrevRepBlockHash []byte, PrevTxBlockHashList [][]byte, CoSignature []byte) *SyncBlock{
	block := &SyncBlock{time.Now().Unix(),PrevRepBlockHash, PrevTxBlockHashList,CoSignature, []byte{}}
	return block

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