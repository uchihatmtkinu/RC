package Reputation

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt32
)

const difficulty = 4


// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	RepBlock  *RepBlock
	Target *big.Int
}


// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *RepBlock) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.RepBlock.PrevRepBlockHash,
			pow.RepBlock.HashRep(),
			pow.RepBlock.HashPrevTxBlockHashes(),
			//IntToHex(pow.RepBlock.Timestamp),
			IntToHex(int64(difficulty)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte, bool) {
	var hashInt big.Int
	var hash [32]byte
	var flag bool
	flag = true
	nonce := 0
	fmt.Println("Mining the RepBlock containing")
	for nonce < maxNonce && flag {
		select {
		case candidateRepBlock:=<-RepPowRxCh:{
			if pow.Validate(candidateRepBlock.Nonce){
				nonce = candidateRepBlock.Nonce
				copy(hash[:], candidateRepBlock.Hash)
				flag = false
				return nonce, hash[:], flag
			}
		}
		default:
			{
				data := pow.prepareData(nonce)
				hash = sha256.Sum256(data)
				fmt.Printf("\r%x", hash)
				hashInt.SetBytes(hash[:])

				if hashInt.Cmp(pow.Target) == -1 {
					return nonce, hash[:], flag
				} else {
					nonce++
				}
			}
		}
	}

	return nonce, hash[:], flag
}

// Validate RepBlock's PoW
func (pow *ProofOfWork) Validate(nonce int) bool {
	var hashInt big.Int

	data := pow.prepareData(nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.Target) == -1

	return isValid
}


