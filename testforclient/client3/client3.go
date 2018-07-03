package main


import (
	"github.com/uchihatmtkinu/RC/testforclient/network"
	"fmt"
	"github.com/uchihatmtkinu/RC/shard"
	"time"
	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/gVar"
)


func main() {
	ID := 3
	totalepoch := 2
	network.IntilizeProcess(ID)
	fmt.Println("test begin")
	go network.StartServer(ID)
	<- network.IntialReadyCh
	close(network.IntialReadyCh)

	fmt.Println("MyGloablID: ", network.MyGlobalID)
	for k:= 1; k<=totalepoch; k++ {
		//test shard
		network.ShardProcess()

		//test rep
		network.RepProcess(&shard.GlobalGroupMems)
		Reputation.CurrentRepBlock.Mu.RLock()
		Reputation.CurrentRepBlock.Block.Print()
		Reputation.CurrentRepBlock.Mu.RUnlock()
		for i := 0; i < int(gVar.ShardSize); i++ {
			shard.GlobalGroupMems[shard.ShardToGlobal[shard.MyMenShard.Shard][i]].AddRep(int64(shard.ShardToGlobal[shard.MyMenShard.Shard][i]))
		}
		network.RepProcess(&shard.GlobalGroupMems)
		Reputation.CurrentRepBlock.Mu.RLock()
		Reputation.CurrentRepBlock.Block.Print()
		Reputation.CurrentRepBlock.Mu.RUnlock()

		//test cosi
		if shard.MyMenShard.Role == shard.RoleLeader {
			network.LeaderCosiProcess(&shard.GlobalGroupMems)
		} else {
			network.MemberCosiProcess(&shard.GlobalGroupMems)
		}

		//test sync
		network.SyncProcess(&shard.GlobalGroupMems)
		time.Sleep(10*time.Second)


		Reputation.CurrentSyncBlock.Mu.RLock()
		Reputation.CurrentSyncBlock.Block.Print()
		Reputation.CurrentSyncBlock.Mu.RUnlock()

		for i:=0 ; i < int(gVar.ShardSize*gVar.ShardCnt); i++ {
			shard.GlobalGroupMems[i].Print()
		}

		Reputation.CurrentSyncBlock.Mu.RLock()
		Reputation.CurrentSyncBlock.Block.Print()
		Reputation.CurrentSyncBlock.Mu.RUnlock()

		for i:=0 ; i < int(gVar.ShardSize*gVar.ShardCnt); i++ {
			shard.GlobalGroupMems[i].Print()
			fmt.Println()
		}
	}

	fmt.Println("All finished")



	time.Sleep(600*time.Second)
}

