package main

import (
	"fmt"
	"time"
)

const n = 5

func main() {
	count := 0
	for count < 3 {
		select {
		case <-time.After(2 * time.Second):
			fmt.Println("timeout 2")
			count++
		}
	}

}
