package basic

import (
	"crypto/ecdsa"
	"math/big"
)

//Miner is the miner
type Miner struct {
	ID        string
	Rep       int
	Prk       ecdsa.PublicKey
	LastGroup int
}

//OutType is the format of the output address data in the transaction
type OutType struct {
	Value   uint32
	Address [32]byte
}

//InType is the format of the input address data in the transaction
type InType struct {
	PrevTx [32]byte
	Index  uint32
	SignR  *big.Int
	SignS  *big.Int
	PrkX   *big.Int
	PrkY   *big.Int
	Acc    bool
}

//RawTransaction is the transaction data which sent by the sender
type RawTransaction struct {
	Timestamp int64
	TxinCnt   uint32
	In        []InType
	TxoutCnt  uint32
	Out       []OutType
	Kind      uint32
	Locktime  uint32
	Hash      [32]byte
}

//TxBlock introduce the struct of the transaction block
type TxBlock struct {
	PrevHash   [32]byte
	TxCnt      uint32
	TxArray    []RawTransaction
	Timestamp  int64
	Height     uint32
	SignR      *big.Int
	SignS      *big.Int
	PrkX       *big.Int
	PrkY       *big.Int
	HashID     [32]byte
	MerkleRoot [32]byte
}

//TxDB is the database of cache
type TxDB struct {
	Data RawTransaction
	Used []uint32
	Res  int8
}

//UserClient is the struct for miner and client
type UserClient struct {
	IPaddress string
	Prk       ecdsa.PublicKey
	kind      int
}
