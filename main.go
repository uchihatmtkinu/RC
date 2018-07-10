package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
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
	network.IntilizeProcess(string(buffer), &ID, os.Args[2])
	fmt.Println("test begin")
	go network.StartServer(ID)
	<-network.IntialReadyCh
	close(network.IntialReadyCh)

	fmt.Println("MyGloablID: ", network.MyGlobalID)
	numCnt := int(gVar.ShardCnt * gVar.ShardSize)
	tmptx := make([]basic.Transaction, gVar.NumOfTxForTest)
	//cnt := 0
	rand.Seed(0)

	for k := 1; k <= totalepoch; k++ {
		//test shard

		network.ShardProcess()
		for l := 0; l < len(tmptx); l++ {
			i := rand.Int() % int(gVar.ShardSize)
			j := rand.Int() % numCnt
			tmptx[l] = *rccache.GenerateTx(shard.ShardToGlobal[network.CacheDbRef.ShardNum][i], j, 1)
		}
		gVar.T1 = time.Now()
		if shard.MyMenShard.Role == shard.RoleLeader {
			go network.SendLoop(&tmptx)
		}
		if k == 1 {
			go network.TxGeneralLoop()
		}
		//test rep
		go network.RepProcessLoop(&shard.GlobalGroupMems)
		//Reputation.CurrentRepBlock.Mu.RLock()
		//Reputation.CurrentRepBlock.Block.Print()
		//Reputation.CurrentRepBlock.Mu.RUnlock()
		/*for i := 0; i < int(gVar.ShardSize); i++ {
			shard.GlobalGroupMems[shard.ShardToGlobal[shard.MyMenShard.Shard][i]].AddRep(int64(shard.ShardToGlobal[shard.MyMenShard.Shard][i]))
		}*/

		//test cosi
		if shard.MyMenShard.Role == shard.RoleLeader {
			network.LeaderCosiProcess(&shard.GlobalGroupMems)
		} else {
			network.MemberCosiProcess(&shard.GlobalGroupMems)
		}

		//test sync
		network.SyncProcess(&shard.GlobalGroupMems)

		/*Reputation.CurrentSyncBlock.Mu.RLock()
		Reputation.CurrentSyncBlock.Block.Print()
		Reputation.CurrentSyncBlock.Mu.RUnlock()
		network.CacheDbRef.Mu.Lock()
		fmt.Println("FB from", network.CacheDbRef.ID)
		for i := uint32(0); i < gVar.ShardCnt; i++ {
			network.CacheDbRef.FB[i].Print()
		}
		network.CacheDbRef.Mu.Unlock()

		for i := 0; i < int(gVar.ShardSize*gVar.ShardCnt); i++ {
			shard.GlobalGroupMems[i].Print()
		}*/

	}

	fmt.Println(network.CacheDbRef.ID, ": All finished")

	//time.Sleep(20 * time.Second)

}
