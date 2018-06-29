package main

import (
	"github.com/uchihatmtkinu/RC/testforclient/network"
	"fmt"
	"github.com/uchihatmtkinu/RC/shard"
)

func main() {
	fmt.Println("test begin")
	network.IntilizeProcess(1)
	fmt.Println(network.MyGlobalID)
	for i, it := range shard.GlobalGroupMems {
		fmt.Println()
		it.AddRep(int64(i))
		it.Print()
	}
	network.ShardProcess()
}
