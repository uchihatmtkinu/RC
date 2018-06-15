package main

import (
	"time"
	"fmt"
)

var a chan bool
func fun(){
	time.Sleep(5*time.Second)
	a <- true
}

func main() {
	a = make(chan bool)
	go fun()
	if <-a {
		fmt.Print("Yes")
	}
	fmt.Print("sorry")
}
