package main

import (
	"fmt"
)

const n = 5

func main() {
	var a [][]int
	var b [][n]int
	//a = make([][]int, 0)
	b = make([][n]int, 3)
	for i :=0; i < 3; i++{
		tmp := make([]int, n)
		for j := 0; j< n; j++{
			tmp[j] = j
		}
		a = append(a, tmp)
		copy(b[i][:],tmp)
	}

	fmt.Println(a)
	fmt.Println(b)
	var c []int
	c = []int{0, 1,2,3,4,5,6}
	c = a[1]
	fmt.Println(c)
}
