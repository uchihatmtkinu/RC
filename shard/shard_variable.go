package shard

//ShardToGlobal shard ind+in shard ind -> global index
var ShardToGlobal [][]int

//GlobalGroupMems global memshard
var GlobalGroupMems []MemShard

//NumMems number of members within one shard
var NumMems int

//MyMenShard my
var MyMenShard *MemShard


//PreviousSyncBlockHash the hash array of previous sync block from all the shards
var PreviousSyncBlockHash [][32]byte

//StartFlag indicate whether it is the first block generated in this epoch
var StartFlag bool

