package main

import (
"github.com/uchihatmtkinu/RC/testforclient/network"
"fmt"
"github.com/uchihatmtkinu/RC/shard"
	"time"
)


func main() {
	network.IntialReadyCh = make(chan bool)
	//totalNum := int(gVar.ShardSize*gVar.ShardCnt)
	fmt.Println("test begin")
	go network.StartServer(1)
	<- network.IntialReadyCh
	close(network.IntialReadyCh)

	fmt.Println("MyGloablID: ", network.MyGlobalID)

	network.ShardProcess()
	/*
	for i:=0 ; i < totalNum; i++ {
		fmt.Println()
		shard.GlobalGroupMems[i].AddRep(int64(i))
		shard.GlobalGroupMems[i].SetTotalRep(int64(i))
		shard.GlobalGroupMems[i].SetTotalRep(int64(i+1))
		shard.GlobalGroupMems[i].SetTotalRep(int64(i+2))
		shard.GlobalGroupMems[i].Print()
		fmt.Println(shard.GlobalGroupMems[i].CalTotalRep())
	}*/
	time.Sleep(5*time.Second)
	network.RepProcess(&shard.GlobalGroupMems)
}
