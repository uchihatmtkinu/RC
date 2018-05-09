package shard

//MemShard is the struct of miners for sharding and leader selection
type MemShard struct {
	address string
	rep     int
	shard   int
}
