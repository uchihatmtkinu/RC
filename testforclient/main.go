package main

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/account"
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/testforclient/network"
)

var MyAccount account.RcAcc
var MyMenShard shard.MemShard
var MyRepBlockChain Reputation.RepBlockchain

func main() {
	cli := network.CLI{}
	fmt.Println("input port")
	var input string
	fmt.Scanln(&input)
	fmt.Println("input height")
	var i int
	fmt.Scanln(&i)
	cli.StartNode(input, i)
}
