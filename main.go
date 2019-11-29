package main

import (
	"bufio"
	"fmt"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/rccache"
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/testforclient/network"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func main() {
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

	scannerMyIP := bufio.NewScanner(file)
	scannerMyIP.Split(bufio.ScanWords)
	scannerMyIP.Scan()
	fmt.Println("Local Address:", scannerMyIP.Text())

	ID := 0
	totalepoch := 1
	network.IntilizeProcess(scannerMyIP.Text(), &ID, os.Args[2], initType)
	//network.IntilizeProcess("192.168.108.37", &ID, os.Args[2], initType)
	go network.StartServer(ID)
	<-network.IntialReadyCh
	close(network.IntialReadyCh)
	// ID merkle tree generation
	network.IDMerkleTreeProcess()
	// update id
	network.IDUpdateProcess()

	fmt.Println("MyGloablID: ", network.MyGlobalID)

	numCnt := int(gVar.ShardCnt * gVar.ShardSize)
	tmptx := make([]basic.Transaction, gVar.NumOfTxForTest*gVar.NumTxListPerEpoch)

	time.Sleep(time.Second * 20)
	timestart := time.Now()
	fmt.Println(time.Now(), "test begin")
	//network.StopChan = make(chan os.Signal, 1)
	for k := 1; k <= totalepoch; k++ {
		//test shard
		fmt.Println("Current time: ", time.Now())
		network.ShardProcess()

		rand.Seed(int64(network.CacheDbRef.ID*3000) + time.Now().Unix()%3000)
		for l := 0; l < len(tmptx); l++ {
			i := rand.Int() % numCnt
			j := rand.Int() % numCnt
			k := uint32(1)
			tmptx[l] = *rccache.GenerateTx(i, j, k, rand.Int63(), network.CacheDbRef.ID+uint32(l*2000))
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
			//if shard.MyMenShard.InShardId < 50 {
			go network.SendLoopMiner(&tmptx)
			//}
			go network.HandleTx()
		}
		//test rep
		//go network.RepProcessLoop(&shard.GlobalGroupMems)
		<-network.RepFinishChan[gVar.NumberRepPerEpoch-1]
		//test cosi
		if network.CacheDbRef.Leader == network.CacheDbRef.ID {
			network.LeaderCosiProcess(&shard.GlobalGroupMems)
		} else {
			network.MemberCosiProcess(&shard.GlobalGroupMems)
		}

		//test sync
		network.SyncProcess(&shard.GlobalGroupMems)
		fmt.Println(time.Now(), "Epoch", k, "finished")

	}

	fmt.Println("Time: ", time.Since(timestart), "TPS:", float64(uint32(totalepoch)*(1+gVar.NumTxListPerEpoch*(gVar.ShardSize-1))*gVar.NumOfTxForTest)/time.Since(timestart).Seconds())
	fmt.Println(network.CacheDbRef.ID, ": All finished")
	if network.CacheDbRef.ID == 0 {
		tmpStr := fmt.Sprint("All finished")
		network.SendTxMessage(gVar.MyAddress, "LogInfo", []byte(tmpStr))
	}

	time.Sleep(20 * time.Second)

	fmt.Println("Total Epoch:", totalepoch, ". ALL DONE, shut down.")
}
