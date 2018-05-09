package shard

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestSortRep(t *testing.T) {
	var a []MemShard
	c := 400000
	for i := 0; i < c; i++ {
		var tmp MemShard
		tmp.address = strconv.Itoa(i)
		tmp.rep = rand.Int() & 100
		a = append(a, tmp)
	}
	b := a[:]
	SortRep(&b, 0, len(b)-1)
	for i := 0; i < len(b)-1; i++ {
		if b[i].rep < b[i+1].rep {
			t.Error(`Sort error`, b[i].rep, b[i+1].rep)
		}
	}
}
