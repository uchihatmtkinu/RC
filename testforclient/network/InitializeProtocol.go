package network

import (
	"crypto/x509"
	"os"
	"strconv"

	"github.com/uchihatmtkinu/RC/account"
	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/cryptonew"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

func intilizeProcess(ID int) {
	currentEpoch = 0
	shard.MyGlobalID = ID
	numCnt := gVar.ShardCnt * gVar.ShardSize
	acc := make([]account.RcAcc, numCnt)
	shard.GlobalGroupMems = make([]shard.MemShard, numCnt)
	file, _ := os.Open("PriKeys.txt")
	accWallet := make([]basic.AccCache, numCnt)
	for i := 0; i < int(numCnt); i++ {
		acc[i].New(strconv.Itoa(i))
		acc[i].NewCosi()
		tmp1 := make([]byte, 121)
		tmp2 := make([]byte, 64)
		file.Read(tmp1)
		xxx, _ := x509.ParseECPrivateKey(tmp1)
		acc[i].Pri = *xxx
		acc[i].Puk = acc[i].Pri.PublicKey
		acc[i].CosiPri = tmp2
		acc[i].CosiPuk = tmp2[32:]
		acc[i].AddrReal = cryptonew.AddressGenerate(&acc[i].Pri)
		acc[i].Addr = base58.Encode(acc[i].AddrReal[:])
		accWallet[i].ID = acc[i].AddrReal
		accWallet[i].Value = 100
		//tmp, _ := x509.MarshalECPrivateKey(&acc[i].Pri)
		shard.GlobalGroupMems[i].NewMemShard(&acc[i])
		//dbs[i].New(uint32(i), acc[i].Pri)
	}
	shard.MyMenShard = shard.GlobalGroupMems[ID]
	CacheDbRef.New(uint32(ID), acc[ID].Pri)
}
