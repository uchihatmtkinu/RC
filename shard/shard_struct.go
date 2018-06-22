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
func (ms *MemShard) newMemShard(acc *account.RcAcc) {
	ms.Address = acc.Addr
	ms.RealAccount = acc
	ms.CosiPub = acc.CosiPuk
	ms.Legal = 0
	ms.Rep = 0
}

//newTotalRep set a new total rep to 0
func (ms *MemShard) newTotalRep() {
	ms.TotalRep = []int64{0}
}
//setInShardId set in shard id
func (ms *MemShard) setInShardId(id int) {
	ms.InShardId = id
}

//setRole 0 - member, 1 - leader
func (ms *MemShard) setRole(role byte) {
	ms.Role = role
}

//setShard set shard
func (ms *MemShard) setShard(shard int) {
	ms.Shard = shard
}
//copyTotalRepFromSB copy total rep from sync bock
func (ms *MemShard) copyTotalRepFromSB(value []int64) {
	ms.TotalRep = value
}
//setTotalRep set totalrep
func (ms *MemShard) setTotalRep(value int64) {
	if len(ms.TotalRep) == gVar.SlidingWindows {
		ms.TotalRep = ms.TotalRep[1:]
	}
	ms.TotalRep = append(ms.TotalRep, value)
}


//addReputation add a reputation value
func (ms *MemShard) addRep(addRep int64) {
	ms.Rep += addRep
}

//calReputation cal total rep over epoches
func (ms *MemShard) calTotalRep() int64 {
	sum := int64(0)
	for i:=range ms.TotalRep {
		sum += ms.TotalRep[i]
	}
	return sum
}

//clearRep clear rep
func (ms *MemShard) clearRep() {
	ms.Rep = 0
}
