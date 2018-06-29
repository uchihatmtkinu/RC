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
	"github.com/uchihatmtkinu/RC/Reputation"
)

func intilizeProcess(ID int) {

	// IP + port
	var IPAddr string
	//current epoch = 0
	currentEpoch = 0


	numCnt := gVar.ShardCnt * gVar.ShardSize

	acc := make([]account.RcAcc, numCnt)
	shard.GlobalGroupMems = make([]shard.MemShard, numCnt)

	shard.MyGlobalID = ID
	port := int64(19000)
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
		//TODO need modify
		port++
		IPAddr =  "143.89.147.72:"+strconv.FormatInt(port,10)
		shard.GlobalGroupMems[i].NewMemShard(&acc[i], IPAddr)
		//map ip+port -> global ID
		GlobalAddrMapToInd[IPAddr] = i
		//dbs[i].New(uint32(i), acc[i].Pri)
	}

	account.MyAccount = acc[ID]
	shard.MyMenShard = &shard.GlobalGroupMems[ID]
	shard.NumMems = int(gVar.ShardSize)
	CacheDbRef.New(uint32(ID), acc[ID].Pri)
	Reputation.CreateRepBlockchain(acc[ID].Addr)
}
