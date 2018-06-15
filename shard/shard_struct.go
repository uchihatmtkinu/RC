package shard

import (
	"github.com/uchihatmtkinu/RC/account"
	"github.com/uchihatmtkinu/RC/ed25519"
)

//MemShard is the struct of miners for sharding and leader selection
type MemShard struct {
	Address     string //ip+port
	Rep         int64 //rep this epoch
	TotalRep	int64 //rep over several epoch
	CosiPub		ed25519.PublicKey
	Shard       int
	InShardId 	int
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
	ms.TotalRep = 0
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
//addReputation add a reputation value
func (ms *MemShard) addRep(addRep int64) {
	ms.Rep += addRep
}
//addReputation add a reputation value on total rep
func (ms *MemShard) addTotalRep(addRep int64) {
	ms.TotalRep += addRep
}
//clearRep clear rep
func (ms *MemShard) clearRep() {
	ms.Rep = 0
}
