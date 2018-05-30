package shard

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"strings"

	"github.com/uchihatmtkinu/RC/basic"
)

//Instance is the struct for sharding
type Instance struct {
	rng rand.Rand
}

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
func (c *Instance) Sharding(a *[]MemShard, b *[][]int) {
	tail := 0
	if uint32(len(*a))%basic.ShardSize > 0 {
		tail = 1
	}
	//rng.Seed()
	now := 0
	nShard := uint32(len(*a)) / basic.ShardSize
	for i := uint32(len(*a)); i < basic.ShardSize+uint32(tail); i++ {
		check := make([]int, nShard)
		tmp := 0
		final := nShard
		if nShard > uint32(len(*a)-now) {
			final = uint32(len(*a) - now)
		}
		if i == basic.ShardSize {
			for j := uint32(0); j < nShard; j++ {
				(*b)[j][i] = -1
			}
		}
		for uint32(tmp) < final {
			x := uint32((c.rng.Int() ^ (*a)[now].rep)) % nShard
			if check[x] == 0 {
				check[x] = 1
				(*b)[x][i] = now
				now++
				tmp++
			}
		}
	}
	for i := uint32(0); i < nShard; i++ {
		c.LeaderSort(a, b, i)
	}
}

//LeaderSort give the priority of being leader in this round
func (c *Instance) LeaderSort(a *[]MemShard, b *[][]int, xx uint32) {
	tmp := make([]float32, len((*b)[xx]))
	for i := 0; i < len(tmp); i++ {
		tmp[i] = c.rng.Float32() / float32((*a)[(*b)[xx][i]].rep)
	}
	for i := 0; i < len(tmp); i++ {
		for j := i + 1; j < len(tmp); j++ {
			if tmp[i] > tmp[j] {
				y := tmp[i]
				tmp[i] = tmp[j]
				tmp[j] = y
				x := (*b)[xx][i]
				(*b)[xx][i] = (*b)[xx][j]
				(*b)[xx][j] = x
			}
		}
	}
}
