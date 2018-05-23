package Reputation

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
	"math"
	"crypto/sha256"
)

type RepBlock struct {
	Timestamp     	 int64
	RepMatrix	  	 [][]byte
	RepTransactions	 []*RepTransaction
	/* TODO
	PrevSyncRepBlockHash [][]byte
	TODO END*/
	PrevRepBlockHash []byte
	Hash          	 []byte
	Nonce         	 int
}

// NewBlock creates and returns Block
func NewRepBlock( RepMatrix [][]int, PrevRepBlockHash []byte) *RepBlock {
	//Trustiness calculation
	var repTransactions []*RepTransaction
	n := len(RepMatrix[1])
	trustiness := SimCal(RepMatrix)
	for i := 0; i < n; i++{
		repTransactions = append(repTransactions,NewRepTransaction(MineridToInd[i],byte(int(trustiness[i])+128)))
	}
	//convert int to byte; 0 byte = -128 int 128; byte = 0 int 255; byte = 127 int
	RepMatrixInByte := make([][]byte, n*n)
	for i := 0; i < n; i++{
		RepMatrixInByte[i] = make([]byte, n)
		for j := 0; j < n; j++{
			RepMatrixInByte[i] = append(RepMatrixInByte[i],byte(RepMatrix[i][j]+128))
		}
	}
	//generate new block
	block := &RepBlock{time.Now().Unix(), RepMatrixInByte,repTransactions, PrevRepBlockHash, []byte{}, 0}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisRepBlock() *RepBlock {
	return NewRepBlock([][]int{{0}},[]byte{})
}

//RepMatrix to Similarity score TODO
func SimCal(RepMatrix [][]int)[]float32{
	n := len(RepMatrix[1])
	sim := make([][]float32,n)
	for i:= range sim{
		sim[i] = make([]float32,n)
	}
	var tmp float32
	for i:=0; i < n; i++{
		for j:=0; j< n; j++{
			if i!=j {
				tmp = 0
				for k:=0;k<n;k++{
					tmp = tmp + (float32(RepMatrix[i][k])/float32(n)-float32(RepMatrix[j][k])/float32(n))*(float32(RepMatrix[i][k])/float32(n)-float32(RepMatrix[j][k])/float32(n))

				}
			}
			sim[i][j] = float32(1.0 - math.Sqrt(float64(tmp/float32(n))))
		}
	}
	sumSim := make([]float32, n)
	minSim := make([]float32, n)
	var tmpmin float32
	var trustiness []float32
	for i:=0; i < n; i++{
		tmp = 0
		tmpmin = math.MaxFloat32
		for j:=0; j< n; j++{
			if i!=j {
				tmp = tmp + sim[j][i]
				if (tmpmin > sim[j][i]){
					tmpmin = sim[j][i]
				}
			}
			sumSim[i] = tmp
			minSim[i] = tmpmin
		}
	}
	for i:=0; i < n; i++{
		tmp = 0
		for j:=0; j< n; j++{
			if i!=j {
				tmp = tmp + (float32(RepMatrix[j][i])-5)*(sim[i][j])/(sumSim[j]-minSim[j])
				}
			trustiness = append(trustiness,tmp)
		}
	}
	return trustiness
}
// HashTransactions returns a hash of the transactions in the block
func (b *RepBlock) HashRepMatrix() []byte {
	var txHashes []byte
	var txHash [32]byte

	for i := range b.RepMatrix {
		for _,value := range b.RepMatrix[i]{
			txHashes = append(txHashes, value)
		}
	}
	txHash = sha256.Sum256(bytes.Join([][]byte{txHashes}, []byte{}))
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