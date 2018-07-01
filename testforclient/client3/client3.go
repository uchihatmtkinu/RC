package main

import (
"github.com/uchihatmtkinu/RC/testforclient/network"
"fmt"
	"github.com/uchihatmtkinu/RC/shard"
)


func main() {
	ID := 3
	network.IntilizeProcess(ID)
	//totalNum := int(gVar.ShardSize*gVar.ShardCnt)
	fmt.Println("test begin")
	go network.StartServer(ID)
	<- network.IntialReadyCh
	close(network.IntialReadyCh)

	fmt.Println("MyGloablID: ", network.MyGlobalID)

	network.ShardProcess()
	//<- network.ShardReadyCh
	//close(network.ShardReadyCh)
	//for i:=0 ; i < totalNum; i++ {
	//	fmt.Println()
	//shard.GlobalGroupMems[i].AddRep(int64(i))
	//shard.GlobalGroupMems[i].SetTotalRep(int64(i))
	//shard.GlobalGroupMems[i].SetTotalRep(int64(i+1))
	//shard.GlobalGroupMems[i].SetTotalRep(int64(i+2))
	//shard.GlobalGroupMems[i].Print()
	//fmt.Println(shard.GlobalGroupMems[i].CalTotalRep())
	//}
	if shard.MyMenShard.Role == shard.RoleLeader {
		network.LeaderCosiProcess(&shard.GlobalGroupMems, [32]byte{1})
	}	else {
		network.MemberCosiProcess(&shard.GlobalGroupMems,[32]byte{1})
	}
	network.SyncProcess(&shard.GlobalGroupMems)
	//<- network.SyncReadyCh
	//close(network.SyncReadyCh)
}