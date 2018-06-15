package shard

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"strings"

	"github.com/uchihatmtkinu/RC/basic"
)

type sortType struct {
	ID      uint32
	Rep     uint64
	Address string
}

//Instance is the struct for sharding
type Instance struct {
	rng rand.Rand
}

//GetRBData get all the data from reputation block for sharding
func GetRBData() {

}

//CompareRep returns whether a has a great reputation than b
func CompareRep(a *sortType, b *sortType) int {
	if a.Rep > b.Rep {
		return 1
	} else if b.Rep > a.Rep {
		return -1
	} else {
		return strings.Compare(a.Address, b.Address)
	}
}

//SortRep sorts all miners based on their reputation
func SortRep(a *[]sortType, l int, r int) error {
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
	sortData := make([]sortType, len(*a))
	for i := 0; i < len(*a); i++ {
		sortData[i].Address = (*a)[i].Address
		sortData[i].ID = uint32(i)
		sortData[i].Rep = (*a)[i].Rep
	}
	SortRep(&sortData, 0, len(*a)-1)
	b = new([][]int)
	*b = make([][]int, basic.ShardCnt)
	//rng.Seed()
	now := 0
	for i := uint32(0); i < basic.ShardSize; i++ {
		(*b)[i] = make([]int, basic.ShardSize)
		for j := uint32(0); j < basic.ShardCnt; j++ {
			(*b)[j][i] = -1
		}
	}
	for i := uint32(len(*a)); i < basic.ShardSize; i++ {
		check := make([]int, basic.ShardCnt)
		for j := uint32(0); j < basic.ShardCnt; j++ {
			x := uint32((c.rng.Int() ^ int((*a)[sortData[now].ID].Rep))) % basic.ShardCnt
			if check[x] == 0 {
				check[x] = 1
				(*b)[x][i] = int(sortData[now].ID)
				now++
			} else {
				j--
			}
		}
	}
	for i := uint32(0); i < basic.ShardCnt; i++ {
		c.LeaderSort(a, b, i)
	}
}

//LeaderSort give the priority of being leader in this round
func (c *Instance) LeaderSort(a *[]MemShard, b *[][]int, xx uint32) {
	tmp := make([]float32, len((*b)[xx]))
	for i := 0; i < len(tmp); i++ {
		tmp[i] = c.rng.Float32() / float32((*a)[(*b)[xx][i]].Rep)
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
