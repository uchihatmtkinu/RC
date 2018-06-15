package main

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/account"
	"github.com/uchihatmtkinu/RC/testforclient/network"
)





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
