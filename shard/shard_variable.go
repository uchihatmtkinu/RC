package shard

//shard ind+in shard ind -> global index
var ShardToGlobal [][]int

var GlobalGroupMems []MemShard

//number of members within one shard
var NumMems int

//my
var MyMenShard MemShard