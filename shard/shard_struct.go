package shard

import (
	"github.com/uchihatmtkinu/RC/account"
	"github.com/uchihatmtkinu/RC/ed25519"
)

//MemShard is the struct of miners for sharding and leader selection
type MemShard struct {
	Address     string
	Rep         uint64
	CosiPub		ed25519.PublicKey
	Shard       int
	InShardId 	int
	Role        byte //0 - member, 1 - leader
	Legal       byte //0 - legal,  1 - kickout
	RealAccount *account.RcAcc
}

func (ms *MemShard) newMemShard(acc *account.RcAcc) {
	ms.Address = acc.Addr
	ms.RealAccount = acc
	ms.CosiPub = acc.CosiPuk
	ms.Legal = 0
	ms.Rep = 0
}

func (ms *MemShard) setInShardId(id int) {
	ms.InShardId = id
}
//0 - member, 1 - leader
func (ms *MemShard) setRole(role byte) {
	ms.Role = role
}

func (ms *MemShard) setShard(shard int) {
	ms.Shard = shard
}

func (ms *MemShard) addReputation(addRep uint64) {
	ms.Rep += addRep
}

func (ms *MemShard) resetRep() {
	ms.Rep = 0
}
