package main


import (
"github.com/uchihatmtkinu/RC/testforclient/network"
"fmt"
"github.com/uchihatmtkinu/RC/shard"
)


func main() {
	network.IntialReadyCh = make(chan bool)
	//totalNum := int(gVar.ShardSize*gVar.ShardCnt)
	fmt.Println("test begin")
	go network.StartServer(3)
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
		//shard.GlobalGroupMems[i].Print()
		//fmt.Println(shard.GlobalGroupMems[i].CalTotalRep())
	}*/
	network.RepProcess(&shard.GlobalGroupMems)
}

