package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/testforclient/network"
)

func main() {
	go network.StartLocalServer()
	PubfileIP, err := os.Open("IPp3.txt")
	defer PubfileIP.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var tmp int
	flag := true
	for flag {
		fmt.Scanln(&tmp)
		if tmp == 1 {
			flag = false
		}
	}

	scannerPub := bufio.NewScanner(PubfileIP)
	scannerPub.Split(bufio.ScanWords)

	IPCnt := gVar.ShardCnt * gVar.ShardSize / 2
	for i := 0; i < int(IPCnt); i++ {

		scannerPub.Scan()
		IPAddrPub := scannerPub.Text()

		IPAddr2 := IPAddrPub + ":" + strconv.Itoa(3000+i)
		network.SendTxMessage(IPAddr2, "shutDown", []byte(""))
		IPAddr2 = IPAddrPub + ":" + strconv.Itoa(3000+i+int(IPCnt))
		network.SendTxMessage(IPAddr2, "shutDown", []byte(""))
	}
	fmt.Println("all shut down")
}
