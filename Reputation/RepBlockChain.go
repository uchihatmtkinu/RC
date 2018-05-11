package Reputation

import (

	"github.com/boltdb/bolt"
	"log"
)

const dbFile = "RepBlockchain.db"
const blocksBucket = "blocks"

type RepBlockchain struct {
	Tip []byte
	Db *bolt.DB
}

// RepBlockchainIterator is used to iterate over Repblockchain blocks
type RepBlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}


// AddBlock saves provided data as a block in the Repblockchain
func (bc *RepBlockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := NewRepBlock(data, lastHash)

	err = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.Tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
// NewBlockchain creates a new Blockchain with genesis Block
func NewRepBlockchain() *RepBlockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := NewGenesisRepBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := RepBlockchain{tip, db}
	return &bc
}
func (bc *RepBlockchain) Iterator() *RepBlockchainIterator {
	bci := &RepBlockchainIterator{bc.Tip, bc.Db}

	return bci
}

func (i *RepBlockchainIterator) Next() *RepBlock {
	var block *RepBlock

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	i.currentHash = block.PrevRepBlockHash

	return block
}