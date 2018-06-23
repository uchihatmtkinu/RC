package Reputation

import (

	"github.com/boltdb/bolt"
	"log"
	"os"
	"fmt"
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/rccache"
	"strconv"
	"github.com/uchihatmtkinu/RC/gVar"
)

const dbFile = "RepBlockchain.db"
const blocksBucket = "blocks"

//reputation block chain
type RepBlockchain struct {
	Tip [32]byte
	Db *bolt.DB
}

// RepBlockchainIterator is used to iterate over Repblockchain blocks
type RepBlockchainIterator struct {
	currentHash [32]byte
	db          *bolt.DB
}

// MineRepBlock mines a new repblock with the provided transactions
func (bc *RepBlockchain) MineRepBlock(ms *[]shard.MemShard, cache *rccache.DbRef) {
	var lastHash [32]byte

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		copy(lastHash[:], b.Get([]byte("lb")))
		//lastHash = b.Get([]byte("lb"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newRepBlock := NewRepBlock(ms, false,  cache.TBCache ,lastHash)
	cache.TBCache = nil
	err = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newRepBlock.Hash[:], newRepBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("lb"), newRepBlock.Hash[:])
		if err != nil {
			log.Panic(err)
		}

		bc.Tip = newRepBlock.Hash

		return nil
	})
}

// add a new syncBlock on RepBlockChain
func (bc *RepBlockchain) AddSyncBlock(ms *[]shard.MemShard, CoSignature []byte) {
	var lastRepBlockHash [32]byte
	var prevSyncBlockHash [][32]byte
	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		copy(lastRepBlockHash[:], b.Get([]byte("lsb")))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	prevSyncBlockHash = make([][32]byte, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		err = bc.Db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			copy(prevSyncBlockHash[i][:], b.Get([]byte("lsb"+strconv.FormatInt(int64(i), 10))))
			return nil
		})
	}
	if err != nil {
		log.Panic(err)
	}

	newSyncBlock := NewSynBlock(ms, prevSyncBlockHash, lastRepBlockHash,  CoSignature)

	err = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newSyncBlock.Hash[:], newSyncBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("lsb"+strconv.FormatInt(int64(shard.MyMenShard.Shard), 10)), newSyncBlock.Hash[:])
		if err != nil {
			log.Panic(err)
		}

		bc.Tip = newSyncBlock.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

}

//AddSyncBlockFromOtherShards add sync block from k-th shard
func (bc *RepBlockchain) AddSyncBlockFromOtherShards(syncBlock *SyncBlock, k int) {
	err := bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(syncBlock.Hash[:], syncBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("lsb"+strconv.FormatInt(int64(k), 10)), syncBlock.Hash[:])
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}


// NewBlockchain creates a new Blockchain with genesis Block
func NewRepBlockchain(nodeID string) *RepBlockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	var tip [32]byte
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
			err = b.Put(genesis.Hash[:], genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = b.Put([]byte("lb"), genesis.Hash[:])
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			copy(tip[:], b.Get([]byte("lb")))
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bc := RepBlockchain{tip, db}
	return &bc
}


// CreateRepBlockchain creates a new blockchain DB
func CreateRepBlockchain(nodeID string) *RepBlockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip [32]byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		genesis := NewGenesisRepBlock()

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash[:], genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("lb"), genesis.Hash[:])
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := RepBlockchain{tip, db}

	return &bc
}


//iterator
func (bc *RepBlockchain) Iterator() *RepBlockchainIterator {
	bci := &RepBlockchainIterator{bc.Tip, bc.Db}

	return bci
}

// Next returns next block starting from the tip
func (i *RepBlockchainIterator) Next() *RepBlock {
	var block *RepBlock

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash[:])
		block = DeserializeRepBlock(encodedBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	i.currentHash = block.PrevRepBlockHash

	return block
}

// NextSB returns next block starting from the tip
func (i *RepBlockchainIterator) NextFromSB() *SyncBlock {
	var block *SyncBlock

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash[:])
		block = DeserializeSyncBlock(encodedBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	i.currentHash = block.PrevRepBlockHash

	return block
}

func (i *RepBlockchainIterator) NextToStart() *RepBlock {
	var block *RepBlock
	var flag bool
	flag = false
	for !flag {
		err := i.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			encodedBlock := b.Get(i.currentHash[:])
			block = DeserializeRepBlock(encodedBlock)

			return nil
		})
		if err != nil {
			log.Panic(err)
		}
		i.currentHash = block.PrevRepBlockHash
		flag = block.StartBlock
	}
	return block
}


//whether database exisits
func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}