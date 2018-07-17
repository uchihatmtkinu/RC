package network

import (
	"crypto/x509"
	"fmt"
	"os"
	"strconv"

	"bufio"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/account"
	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/cryptonew"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//IntilizeProcess is init
func IntilizeProcess(input string, ID *int, PriIPFile string, PubIPFile string, initType int) {

	// IP + port
	var IPAddrPri, IPAddrPub string
	fmt.Println("Initlization:", input, PriIPFile, initType)

	numCnt := gVar.ShardCnt * gVar.ShardSize

	acc := make([]account.RcAcc, numCnt)
	shard.GlobalGroupMems = make([]shard.MemShard, numCnt)
	//GlobalAddrMapToInd = make(map[string]int)

	file, err := os.Open("PriKeys.txt")
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	PubfileIP, err := os.Open(PubIPFile)
	defer PubfileIP.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	PrifileIP, err := os.Open(PriIPFile)
	defer PrifileIP.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	scannerPub := bufio.NewScanner(PubfileIP)
	scannerPub.Split(bufio.ScanWords)
	scannerPri := bufio.NewScanner(PrifileIP)
	scannerPri.Split(bufio.ScanWords)

	accWallet := make([]basic.AccCache, numCnt)
	for i := 0; i < int(numCnt); i++ {
		//scanner.Scan()
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
		accWallet[i].Value = 100000000
	}
	IPCnt := int(numCnt)
	if initType != 0 {
		IPCnt /= 2
	}
	//tmp, _ := x509.MarshalECPrivateKey(&acc[i].Pri)
	//TODO need modify
	shardRegion := 1
	for i := 0; i < int(IPCnt); i++ {
		scannerPri.Scan()
		scannerPub.Scan()
		IPAddrPri = scannerPri.Text()
		IPAddrPub = scannerPub.Text()

		IPAddr1 := IPAddrPri + ":" + strconv.Itoa(3000+i)
		IPAddr2 := IPAddrPub + ":" + strconv.Itoa(3000+i)
		var band int
		if gVar.BandDiverse {
			band = 2048 + 1024*(i+1)*2/100
		} else {
			band = 1024 * 38
		}
		shard.GlobalGroupMems[i].NewMemShard(&acc[i], IPAddr1, IPAddr2, band)
		shard.GlobalGroupMems[i].NewTotalRep()
		//shard.GlobalGroupMems[i].AddRep(int64(i))
		if initType != 0 {
			IPAddr1 := IPAddrPri + ":" + strconv.Itoa(3000+i+IPCnt)
			IPAddr2 := IPAddrPub + ":" + strconv.Itoa(3000+i+IPCnt)
			if gVar.BandDiverse {
				band = gVar.MinBand + (gVar.MaxBand-gVar.MinBand)*(i+1)/int(numCnt)
			} else {
				band = gVar.MaxBand
			}
			shard.GlobalGroupMems[i+IPCnt].NewMemShard(&acc[i+IPCnt], IPAddr1, IPAddr2, band)
			shard.GlobalGroupMems[i+IPCnt].NewTotalRep()
			//shard.GlobalGroupMems[i+IPCnt].AddRep(int64(i + IPCnt))
		}
		if IPAddrPri == input {
			if i >= IPCnt/2 {
				shardRegion = 2
			}
			MyGlobalID = i
			*ID = i
			if initType == 2 {
				MyGlobalID += IPCnt
				*ID += IPCnt
			}
			bindAddress = ":" + strconv.Itoa(3000+MyGlobalID)
		}
		//map ip+port -> global ID
		//GlobalAddrMapToInd[IPAddr] = i
		//dbs[i].New(uint32(i), acc[i].Pri)
	}
	if shardRegion == 1 {
		for i := 0; i < IPCnt/2; i++ {
			shard.GlobalGroupMems[i].Address = shard.GlobalGroupMems[i].PrivateAddress
			if initType != 0 {
				shard.GlobalGroupMems[i+IPCnt].Address = shard.GlobalGroupMems[i+IPCnt].PrivateAddress
			}
		}
	} else {
		for i := IPCnt / 2; i < IPCnt; i++ {
			shard.GlobalGroupMems[i].Address = shard.GlobalGroupMems[i].PrivateAddress
			if initType != 0 {
				shard.GlobalGroupMems[i+IPCnt].Address = shard.GlobalGroupMems[i+IPCnt].PrivateAddress
			}
		}
	}
	CacheDbRef.New(uint32(*ID), acc[*ID].Pri)
	for i := 0; i < int(numCnt); i++ {
		CacheDbRef.DB.AddAccount(&accWallet[i])
	}
	account.MyAccount = acc[*ID]

	shard.MyMenShard = &shard.GlobalGroupMems[*ID]
	shard.NumMems = int(gVar.ShardSize)
	shard.PreviousSyncBlockHash = [][32]byte{{gVar.MagicNumber}}

	Reputation.RepPowRxCh = make(chan Reputation.RepPowInfo, bufferSize)
	Reputation.CurrentSyncBlock = Reputation.SafeSyncBlock{Block: nil, Epoch: -1}
	Reputation.CurrentRepBlock = Reputation.SafeRepBlock{Block: nil, Round: -1}
	Reputation.MyRepBlockChain = Reputation.CreateRepBlockchain(strconv.FormatInt(int64(MyGlobalID), 10))

	//current epoch = -1
	CurrentEpoch = -1

	//make channel
	IntialReadyCh = make(chan bool)
	ShardReadyCh = make(chan bool)
	CoSiReadyCh = make(chan bool)
	SyncReadyCh = make(chan bool)

	FinalTxReadyCh = make(chan bool, 1)
	waitForFB = make(chan bool, 1)
	//channel used in shard
	readyMemberCh = make(chan readyInfo, bufferSize)
	readyLeaderCh = make(chan readyInfo, bufferSize)
	//channel used in CoSi
	cosiAnnounceCh = make(chan []byte)

	//channel used in final block
	finalSignal = make(chan []byte)
	startRep = make(chan repInfo, 1)
	startSync = make(chan bool, 1)
	StartLastTxBlock = make(chan bool, 1)
	StartNewTxlist = make(chan bool, 1)
	StartSendingTx = make(chan bool, 1)
	for i := uint32(0); i < gVar.NumTxListPerEpoch; i++ {
		TxDecRevChan[i] = make(chan txDecRev, gVar.ShardCnt)
		TLChan[i] = make(chan uint32, gVar.ShardSize)
		txMCh[i] = make(chan txDecRev, gVar.ShardCnt)
	}

}
