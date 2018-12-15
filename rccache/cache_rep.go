package rccache

import (
	"fmt"

	newrep "github.com/uchihatmtkinu/RC/NewRep"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//MakeRepMsg generate the newest RepMsg
func (d *DbRef) MakeRepMsg(round uint32) newrep.RepMsg {
	mydata := new(newrep.RepMsg)
	mydata.Make(d.ID, *d.TBCache, d.RepVote[round][:], round, &d.prk)
	d.RepFirMsg[round][shard.MyMenShard.InShardId] = *mydata
	return *mydata
}

//MakeRepSecMsg generate the newest RepMsg
func (d *DbRef) MakeRepSecMsg(round uint32, g newrep.GossipFirMsg) newrep.RepSecMsg {
	mydata := new(newrep.RepSecMsg)
	mydata.Make(d.ID, g, round, &d.prk)
	d.RepSecMsg[round][shard.MyMenShard.InShardId] = *mydata
	return *mydata
}

//GenerateGossipFir gives the data for gossip
func (d *DbRef) GenerateGossipFir(round uint32) (*newrep.GossipFirMsg, int) {
	d.MakeRepMsg(round)
	tmp := new(newrep.GossipFirMsg)
	tmp.ID = d.ID
	tmp.Cnt = 0
	tmp.Data = nil
	cnt := 0
	tmpArr := make([]uint32, 0, gVar.ShardSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		if d.RepFirSig[round][i] {
			tmp.Add(d.RepFirMsg[round][i])
		} else {
			tmpArr = append(tmpArr, i)
		}
	}
	if cnt > 0 {
		ran := d.GetGossipID(1, tmpArr)
		return tmp, shard.ShardToGlobal[d.ShardNum][ran]
	}
	return nil, 0

}

//GenerateGossipSec gives the data for gossip
func (d *DbRef) GenerateGossipSec(round uint32) (*newrep.GossipSecMsg, int) {
	tmpGossip := new(newrep.GossipFirMsg)
	tmpGossip.ID = d.ID
	tmpGossip.Cnt = 0
	tmpGossip.Data = nil
	for i := uint32(0); i < gVar.ShardSize; i++ {
		if d.RepFirSig[round][i] {
			tmpGossip.Add(d.RepFirMsg[round][i])
		}
	}
	d.MakeRepSecMsg(round, *tmpGossip)
	tmp := new(newrep.GossipSecMsg)
	tmp.ID = d.ID
	tmp.Cnt = 0
	tmp.Data = nil
	cnt := 0
	tmpArr := make([]uint32, 0, gVar.ShardSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		if d.RepSecSig[round][i] {
			tmp.Add(d.RepSecMsg[round][i])
		} else {
			tmpArr = append(tmpArr, i)
		}
	}
	if cnt > 0 {
		ran := d.GetGossipID(1, tmpArr)
		return tmp, shard.ShardToGlobal[d.ShardNum][ran]
	}
	return nil, 0
}

//UpdateGossipFir with the incoming data
func (d *DbRef) UpdateGossipFir(data newrep.GossipFirMsg) newrep.GossipFirMsg {
	d.RepVote[d.RepRound][data.ID].Rep++
	tmpRound := data.Data[0].Round
	tmp := new(newrep.GossipFirMsg)
	tmp.ID = d.ID
	tmp.Cnt = 0
	tmp.Data = nil
	if tmpRound > gVar.NumNewRep {
		fmt.Println("Data round over limit")
		return *tmp
	}
	tmpArr := make([]bool, gVar.ShardSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		tmpArr[i] = d.RepFirSig[tmpRound][i]
	}
	for i := uint32(0); i < data.Cnt; i++ {
		tmpInShardID := shard.GlobalGroupMems[data.Data[i].ID].InShardId
		tmpArr[tmpInShardID] = false
		if d.RepFirSig[tmpRound][tmpInShardID] {
			if d.RepFirMsg[tmpRound][tmpInShardID].Hash() != data.Data[i].Hash() {
				d.RepByz[tmpRound][tmpInShardID] = true
			}
		} else {
			d.RepFirSig[tmpRound][tmpInShardID] = true
			d.RepFirMsg[tmpRound][tmpInShardID] = data.Data[i]
		}
	}
	for i := uint32(0); i < data.Cnt; i++ {
		if tmpArr[i] {
			tmp.Add(d.RepFirMsg[tmpRound][i])
		}
	}
	return *tmp
}

//UpdateGossipSec with the incoming data
func (d *DbRef) UpdateGossipSec(data newrep.GossipSecMsg) newrep.GossipSecMsg {
	d.RepVote[d.RepRound][data.ID].Rep++
	tmpRound := data.Data[0].Round
	tmp := new(newrep.GossipSecMsg)
	tmp.ID = d.ID
	tmp.Cnt = 0
	tmp.Data = nil
	if tmpRound > gVar.NumNewRep {
		fmt.Println("Data round over limit")
		return *tmp
	}

	tmpArr := make([]bool, gVar.ShardSize)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		tmpArr[i] = d.RepSecSig[tmpRound][i]
	}
	for i := uint32(0); i < data.Cnt; i++ {
		tmpInShardID := shard.GlobalGroupMems[data.Data[i].ID].InShardId
		tmpArr[tmpInShardID] = false
		if !d.RepSecSig[tmpRound][tmpInShardID] {
			d.RepSecSig[tmpRound][tmpInShardID] = true
			d.RepSecMsg[tmpRound][tmpInShardID] = data.Data[i]
		}
	}
	for i := uint32(0); i < data.Cnt; i++ {
		if tmpArr[i] {
			tmp.Add(d.RepSecMsg[tmpRound][i])
		}
	}
	return *tmp
}
