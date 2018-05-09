package shard

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"strings"

	"github.com/uchihatmtkinu/RC/basic"
)

var rng rand.Rand

//GetRBData get all the data from reputation block for sharding
func GetRBData() {

}

//CompareRep returns whether a has a great reputation than b
func CompareRep(a *MemShard, b *MemShard) int {
	if a.rep > b.rep {
		return 1
	} else if b.rep > a.rep {
		return -1
	} else {
		return strings.Compare(a.address, b.address)
	}
}

//SortRep sorts all miners based on their reputation
func SortRep(a *[]MemShard, l int, r int) error {
	x := (*a)[(l+r)/2]
	i := l
	j := r
	if l >= r {
		return nil
	}
	for i <= j {
		for i < r && CompareRep(&(*a)[i], &x) > 0 {
			i++
		}
		for j > 0 && CompareRep(&x, &(*a)[j]) > 0 {
			j--
		}
		if i <= j {
			y := (*a)[i]
			(*a)[i] = (*a)[j]
			(*a)[j] = y
			i++
			j--
		}
	}
	if i < r {
		SortRep(a, i, r)
	}
	if l < j {
		SortRep(a, l, j)
	}
	return nil
}

//GenerateSeed come out the seed used in random number
func GenerateSeed(a *[][32]byte) int64 {
	var tmp []byte
	for i := 0; i < len(*a); i++ {
		tmp = append(tmp, (*a)[i][:]...)
	}
	hash := sha256.Sum256(tmp)

	return int64(binary.BigEndian.Uint64(hash[:]))
}

//Sharding do the shards given reputations
func Sharding(a *[]MemShard, b *[][]int) {
	tail := 0
	if len(*a)%basic.ShardSize > 0 {
		tail = 1
	}
	//rng.Seed()
	now := 0
	nShard := len(*a) / basic.ShardSize
	for i := 0; i < basic.ShardSize+tail; i++ {
		check := make([]int, nShard)
		tmp := 0
		final := nShard
		if nShard > (len(*a) - now) {
			final = len(*a) - now
		}
		if i == basic.ShardSize {
			for j := 0; j < nShard; j++ {
				(*b)[j][i] = -1
			}
		}
		for tmp < final {
			x := (rng.Int() ^ (*a)[now].rep) % nShard
			if check[x] == 0 {
				check[x] = 1
				(*b)[x][i] = now
				now++
				tmp++
			}
		}

	}
}

//LeaderSort give the priority of being leader in this round
func LeaderSort(a *[]MemShard, b *[]int) {
	tmp := make([]float32, len(*b))
	index := make([]int, len(*b))
	for i := 0; i < len(tmp); i++ {
		tmp[i] = rng.Float32() / float32((*a)[(*b)[i]].rep)
		index[i] = i
	}
	for i := 0; i < len(tmp); i++ {
		for j := i + 1; j < len(tmp); j++ {
			if tmp[i] > tmp[j] {
				y := tmp[i]
				tmp[i] = tmp[j]
				tmp[j] = y
				x := index[i]
				index[i] = index[j]
				index[j] = x
			}
		}
	}
}
