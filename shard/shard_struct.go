package shard

import (
	"github.com/uchihatmtkinu/RC/account"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/gVar"
)

//MemShard is the struct of miners for sharding and leader selection
type MemShard struct {
	Address     string //ip+port
	Rep         int64  //rep this epoch
	TotalRep    []int64  //rep over several epoch
	CosiPub     ed25519.PublicKey
	Shard       int
	InShardId   int
	Role        byte //1 - member, 0 - leader
	Legal       byte //0 - legal,  1 - kickout
	RealAccount *account.RcAcc
}

//newMemShard new a mem shard
func (ms *MemShard) NewMemShard(acc *account.RcAcc) {
	ms.Address = acc.Addr
	ms.RealAccount = acc
	ms.CosiPub = acc.CosiPuk
	ms.Legal = 0
	ms.Rep = 0
}

//NewTotalRep set a new total rep to 0
func (ms *MemShard) NewTotalRep() {
	ms.TotalRep = []int64{}
}

//CopyTotalRepFromSB copy total rep from sync bock
func (ms *MemShard) CopyTotalRepFromSB(value []int64) {
	ms.TotalRep = value
}
//SetTotalRep set totalrep
func (ms *MemShard) SetTotalRep(value int64) {
	if len(ms.TotalRep) == gVar.SlidingWindows {
		ms.TotalRep = ms.TotalRep[1:]
	}
	ms.TotalRep = append(ms.TotalRep, value)
}


//AddRep add a reputation value
func (ms *MemShard) AddRep(addRep int64) {
	ms.Rep += addRep
}

//CalTotalRep cal total rep over epoches
func (ms *MemShard) CalTotalRep() int64 {
	sum := int64(0)
	for i:=range ms.TotalRep {
		sum += ms.TotalRep[i]
	}
	return sum
}

//ClearRep clear rep
func (ms *MemShard) ClearRep() {
	ms.Rep = 0
}
