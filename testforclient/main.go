package main

import (
	"github.com/uchihatmtkinu/RC/testforclient/network"
	"fmt"
	"github.com/uchihatmtkinu/RC/account"
)

func main() {
	fmt.Println("test begin")
	network.IntilizeProcess(1)
	fmt.Println(account.MyAccount)
}
