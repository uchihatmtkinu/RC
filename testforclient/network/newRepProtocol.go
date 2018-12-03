package network

import (
	"fmt"
	"time"

	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//RepGossipLoop is the loop of New reputation protocol
func RepGossipLoop(ms *[]shard.MemShard) {
	for i := uint32(0); i < gVar.NumNewRep; i++ {
		go NewRepProcess(ms, i)
		time.Sleep(time.Second * 30)
	}
	fmt.Println("Start to sync")
	startSync <- true
}

//NewRepProcess with gossip
func NewRepProcess(ms *[]shard.MemShard, round uint32) {
	//var mydata newrep.RepMsg
	//mydata.Make(uint32(MyGlobalID),[],CacheDbRef.
}
