package rccache

import (
	"bytes"
	"log"
	"math/rand"

	"github.com/uchihatmtkinu/RC/cryptonew"

	"github.com/boltdb/bolt"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/treap"
)

const dbFile = "TxBlockchain.db"

//TXBucket is the bucket of TX
const TXBucket = "TX"

//ACCBucket is the bucket of ACC
const ACCBucket = "ACC"

//byteCompare is the func used for string compare
func byteCompare(a, b interface{}) int {
	switch a.(type) {
	case *basic.AccCache:
		tmp1 := a.(*basic.AccCache).ID
		tmp2 := b.(*basic.AccCache).ID
		return bytes.Compare(tmp1[:], tmp2[:])
	default:
		return bytes.Compare([]byte(a.([]byte)), []byte(b.([]byte)))
	}

}

//dbRef is the structure stores the cache of a miner for the database
type dbRef struct {
	ID       [32]byte
	AccData  *gtreap.Treap
	Mem      *bolt.DB
	TX       map[[32]byte]*basic.CrossShardDec
	tmp      map[[32]byte]*basic.TxDB
	ShardNum uint32
}

//New is the initilization of dbRef
func (d *dbRef) New() {
	d.AccData = gtreap.NewTreap(byteCompare)
	var err error
	d.Mem, err = bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer d.Mem.Close()
	var tmp *basic.AccCache
	err = d.Mem.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(ACCBucket))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			tmp = new(basic.AccCache)
			copy(tmp.ID[:], k[:32])
			tmpStr := v
			basic.DecodeInt(&tmpStr, &tmp.Value)
			d.AccData = d.AccData.Upsert(tmp, rand.Int())
		}

		return nil
	})
	d.TX = make(map[[32]byte]*basic.CrossShardDec)
	d.tmp = make(map[[32]byte]*basic.TxDB)
}

func (d *dbRef) FindACC(ID [32]byte, value uint32) int {
	xxx := basic.AccCache{ID: ID, Value: value}
	tmp := d.AccData.Get(xxx)
	if tmp == nil {
		return -1
	}
	if value > tmp.(*basic.AccCache).Value {
		return -1
	}
	return int(tmp.(*basic.AccCache).Value - value)
}

//FindTX finds the transaction given the index
func (d *dbRef) FindTX(ID [32]byte, index uint32) int {
	var err error
	xxx, ok := d.tmp[ID]
	if !ok {
		if uint32(len(xxx.Used)) <= index {
			return -1
		}
		if xxx.Used[index] == 0 {
			return 1
		}
		return -1
	}
	d.Mem, err = bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
		return -1
	}
	defer d.Mem.Close()
	var res []byte
	err = d.Mem.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TXBucket))
		if b != nil {
			res = b.Get(ID[:])
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
		return -1
	}
	if res == nil {
		return -1
	}
	if err != nil {
		return -1
	}
	tmp := new(basic.TxDB)
	err = (*tmp).Decode(&res)
	d.tmp[ID] = tmp //Check whether it is ok
	if index >= uint32(len(tmp.Used)) {
		return -1
	}
	if tmp.Used[index] == 0 {
		return 1
	}
	return -1
}

//CheckInType deal with the InType in one of the transaction
func (d *dbRef) CheckInType(a basic.InType) int {
	if a.ShardIndex() != d.ShardNum {
		return -2
	}
	if cryptonew.GenerateAddr(a.Puk()) == a.PrevTx {
		return d.FindACC(a.PrevTx, a.Index)
	}
	return d.FindTX(a.PrevTx, a.Index)
}
