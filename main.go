package main

import (
	"fmt"
)

func main() {
	c := make(chan []int,2)

	things := []int{1, 2, 3}
	a1 := []int{2,3}
	go func() {
		c <- things
	}()



	go func() {
		c <- a1
	}()
	a := <-c
	fmt.Println(a)
	fmt.Println(<-c)
}
