package main

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/uchihatmtkinu/RC/gVar"
)

func TestID(t *testing.T) {
	fileIP, err := os.Open("IpAddr.txt")
<<<<<<< HEAD
	ID := 185
=======
	ID := 201
>>>>>>> 6b23e472c30382f8d8540afe668eabc8ee8d2f33

	defer fileIP.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(fileIP)
	scanner.Split(bufio.ScanWords)
	IPCnt := int(gVar.ShardSize * gVar.ShardCnt / 2)
	if ID >= IPCnt {
		ID -= IPCnt
	}
	for i := 0; i < int(IPCnt); i++ {
		scanner.Scan()
		tmp := scanner.Text()
		scanner.Scan()
		IPAddr := scanner.Text()

		if ID == i {
			fmt.Println(IPAddr, tmp)
		}
		//map ip+port -> global ID
		//GlobalAddrMapToInd[IPAddr] = i
		//dbs[i].New(uint32(i), acc[i].Pri)
	}
	t.Error("xx")
}