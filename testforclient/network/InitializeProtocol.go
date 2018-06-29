package network

import (
	"crypto/x509"
	"os"
	"strconv"

	"fmt"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/account"
	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/cryptonew"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

func IntilizeProcess(ID int) {

	// IP + port
	var IPAddr string
	//current epoch = 0
	CurrentEpoch = 0

	numCnt := gVar.ShardCnt * gVar.ShardSize

	acc := make([]account.RcAcc, numCnt)
	shard.GlobalGroupMems = make([]shard.MemShard, numCnt)
	GlobalAddrMapToInd = make(map[string]int)
	MyGlobalID = ID
	port := int64(18999)
	file, err := os.Open("PriKeys.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	accWallet := make([]basic.AccCache, numCnt)
	for i := 0; i < int(numCnt); i++ {
		acc[i].New(strconv.Itoa(i))
		acc[i].NewCosi()
		tmp1 := make([]byte, 121)
		tmp2 := make([]byte, 64)
		file.Read(tmp1)
		file.Read(tmp2)
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
		IPAddr = "192.168.0.108:" + strconv.FormatInt(port, 10)
		shard.GlobalGroupMems[i].NewMemShard(&acc[i], IPAddr)
		shard.GlobalGroupMems[i].NewTotalRep()
		//map ip+port -> global ID
		GlobalAddrMapToInd[IPAddr] = i
		//dbs[i].New(uint32(i), acc[i].Pri)
	}

	account.MyAccount = acc[ID]
	shard.MyMenShard = &shard.GlobalGroupMems[ID]
	shard.NumMems = int(gVar.ShardSize)
	shard.PreviousSyncBlockHash = [][32]byte{{123}}
	CacheDbRef.New(uint32(ID), acc[ID].Pri)
	Reputation.MyRepBlockChain = Reputation.CreateRepBlockchain(strconv.FormatInt(int64(MyGlobalID), 10))
	Reputation.RepPowRxCh = make(chan Reputation.RepBlock, bufferSize)
	cosiAnnounceCh = make(chan []byte)
	cosiChallengeCh = make(chan challengeMessage)
	cosiSigCh = make(chan cosi.SignaturePart)
}
