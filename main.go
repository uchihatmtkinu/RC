package main

import "fmt"

func main() {
	a := []int{1,2,3}
	b := []int{4,5}
	var s []int
	s = append(s,a...)
	fmt.Println(s)
	s = append(s,b...)
	fmt.Println(s)

}
