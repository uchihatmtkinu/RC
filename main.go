package main

import (
	"os"
	"fmt"
	"strconv"
)

func main(){
	var arg int
	//argsWithProg := os.Args
	arg,_ = strconv.Atoi(os.Args[1])

	//fmt.Println(argsWithProg)
	fmt.Println(arg)
}