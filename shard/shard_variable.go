package shard

//ShardToGlobal shard ind+in shard ind -> global index
var ShardToGlobal [][]int

//GlobalGroupMems global memshard
var GlobalGroupMems []MemShard

//NumMems number of members within one shard
var NumMems int

//MyMenShard my
var MyMenShard MemShard

//MyGlobalID my global ID
var MyGlobalID int

//PreviousSyncBlockHash the hash array of previous sync block from all the shards
var PreviousSyncBlockHash [][32]byte


