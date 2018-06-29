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

	network.ShardProcess()
	network.RepProcess(&shard.GlobalGroupMems)
}
