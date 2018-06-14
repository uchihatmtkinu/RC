package main

import (
	"fmt"
	"time"
)


func run(a chan int) {
	time.Sleep(5*time.Second)
	a <- 10

}
func main() {
	a:=make(chan int)
	go run(a)
	for  {
		select {
		case f:=<-a:{
			fmt.Println(f)
			return
		}
		default:
			fmt.Println("N")
		}
	}

}
