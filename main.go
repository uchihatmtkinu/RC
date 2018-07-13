package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/uchihatmtkinu/RC/rccache"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/testforclient/network"
)

func main() {
	//arg, err := strconv.Atoi(os.Args[1])
	/*if err != nil {
		log.Panic(err)
		os.Exit(1)
	}*/
	fmt.Println("Get the local ip from", os.Args[1])
	file, err := os.Open(os.Args[1])
	initType, initErr := strconv.Atoi(os.Args[3])
	if initErr != nil {
		log.Panic(initErr)
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	fileSize := fileinfo.Size()
	buffer := make([]byte, fileSize)

	_, err = file.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Local Address:", string(buffer))

	ID := 0
	totalepoch := 1
	network.IntilizeProcess(string(buffer), &ID, os.Args[2], initType)
	fmt.Println("test begin")
	go network.StartServer(ID)
	<-network.IntialReadyCh
	close(network.IntialReadyCh)

	fmt.Println("MyGloablID: ", network.MyGlobalID)
	numCnt := int(gVar.ShardCnt * gVar.ShardSize)
	tmptx := make([]basic.Transaction, gVar.NumOfTxForTest)
	//cnt := 0
	time.Sleep(time.Second * 20)
	for k := 1; k <= totalepoch; k++ {
		//test shard
		fmt.Println("Current time: ", time.Now())
		network.ShardProcess()
		rand.Seed(int64(network.CacheDbRef.ID) * time.Now().Unix())
		for l := 0; l < len(tmptx); l++ {
			xx := rand.Int() % int(gVar.ShardSize)
			i := shard.ShardToGlobal[network.CacheDbRef.ShardNum][xx]
			j := rand.Int() % numCnt
			k := uint32(rand.Int()%5 + 1)
			tmptx[l] = *rccache.GenerateTx(i, j, k, rand.Int63())
			//fmt.Println(base58.Encode(tmptx[l].Hash[:]))
		}
		gVar.T1 = time.Now()
		fmt.Println("This time", time.Now())

		if shard.MyMenShard.Role == shard.RoleLeader {
			fmt.Println("This is a Leader")
			go network.TxGeneralLoop()
			go network.SendLoopLeader(&tmptx)
			go network.HandleTxLeader()
		} else {
			go network.SendLoopMiner(&tmptx)
			go network.HandleTx()
		}
		//test rep
		go network.RepProcessLoop(&shard.GlobalGroupMems)

		//test cosi
		if shard.MyMenShard.Role == shard.RoleLeader {
			network.LeaderCosiProcess(&shard.GlobalGroupMems)
		} else {
			network.MemberCosiProcess(&shard.GlobalGroupMems)
		}

		//test sync
		network.SyncProcess(&shard.GlobalGroupMems)

	}

	fmt.Println(network.CacheDbRef.ID, ": All finished")

	time.Sleep(20 * time.Second)

}
