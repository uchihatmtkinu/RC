package network

import (
	"fmt"
	"strconv"
	"testing"
)

func TestHighRepAttack(t *testing.T) {
	n := 10
	oldRep := make([]int64, n)
	oldTotalRep := make([][]int64, n)
	oldSumRep := make([]int64, n)
	oldBand := make([]int, n)
	oldAdd := make([]string, n)
	oldSumRep = []int64{1, 1, 2, 2, 3, 3, 4, 5, 6, 7}
	for i := 0; i < n; i++ {
		oldBand[i] = i
		oldRep[i] = int64(i)
		oldAdd[i] = strconv.Itoa(10 - i)
		oldTotalRep[i] = make([]int64, 2)
		for j := 0; j < 2; j++ {
			oldTotalRep[i][j] = int64(i*10 + j)
		}
	}
	AttackSortRep(&oldSumRep, &oldAdd, &oldRep, &oldTotalRep, &oldBand, 0, n-1)
	for i := 0; i < n; i++ {
		fmt.Println(oldSumRep[i], oldAdd[i], oldBand[i])
	}
}
