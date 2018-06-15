package rccache

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
)

//VerifyTx verify the utxos related to transaction a
func (d *DbRef) VerifyTx(a *basic.Transaction) bool {
	for i := uint32(0); i < a.TxinCnt; i++ {
		if a.In[i].ShardIndex() == d.ShardNum {
			if !d.db.CheckUTXO(&a.In[i], a.Hash) {
				return false
			}
		}
	}
	return true
}

//LockTx locks all utxos related to transction a
func (d *DbRef) LockTx(a *basic.Transaction) error {
	d.db.TXCache[a.Hash] = 1
	for i := uint32(0); i < a.TxinCnt; i++ {
		if a.In[i].ShardIndex() == d.ShardNum {
			err := d.db.LockUTXO(&a.In[i])
			if err != nil {
				log.Panic(err)
			}
		}
	}
	return nil
}

//UnlockTx locks all utxos related to transction a
func (d *DbRef) UnlockTx(a *basic.Transaction) error {
	delete(d.db.TXCache, a.Hash)
	for i := uint32(0); i < a.TxinCnt; i++ {
		if a.In[i].ShardIndex() == d.ShardNum {
			err := d.db.UnlockUTXO(&a.In[i])
			if err != nil {
				log.Panic(err)
			}
		}
	}
	return nil
}

//GetTx update the transaction
func (d *DbRef) GetTx(a *basic.Transaction) error {
	tmp, ok := d.TXCache[a.Hash]
	if ok {
		tmp.Update(a)
	} else {
		tmp = new(CrossShardDec)
		tmp.New(a)
		d.TXCache[a.Hash] = tmp
	}
	return nil
}

//ProcessTL verify the transactions in the txlist
func (d *DbRef) ProcessTL(a *basic.TxList) error {
	d.TLNow = new(basic.TxDecision)
	d.TLNow.Set(d.ID, d.ShardNum, 0)
	d.TLNow.HashID = a.HashID
	d.TLNow.Single = 0
	d.TLNow.Sig = make([]basic.RCSign, 0, gVar.ShardCnt)
	var tmpHash [gVar.ShardCnt][]byte
	var tmpDecision [gVar.ShardCnt]basic.TxDecision
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		tmpHash[i] = []byte{}
		tmpDecision[i].Set(d.ID, i, 1)
	}
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp, ok := d.TXCache[a.TxArray[i]]
		if !ok {
			d.TLNow.Add(0)
		} else {
			var res byte
			if tmp.InCheck[d.ShardNum] == 3 {
				if d.VerifyTx(tmp.Data) {
					res = byte(1)
					d.LockTx(tmp.Data)
				}
				d.TLNow.Add(res)
				for j := 0; j < len(tmp.ShardRelated); j++ {
					tmpHash[tmp.ShardRelated[j]] = append(tmpHash[tmp.ShardRelated[j]], a.TxArray[i][:]...)
					tmpDecision[tmp.ShardRelated[j]].Add(res)
				}
			}
		}
	}
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		tmpDecision[i].HashID = sha256.Sum256(tmpHash[i])
		tmpDecision[i].Sign(&d.prk, 0)
		d.TLNow.Sig[i] = tmpDecision[i].Sig[0]
	}
	d.TLSent = d.TLNow
	return nil
}

//GetTDS and ready to verify txblock
func (d *DbRef) GetTDS(b *basic.TxDecSet) error {
	index := 0
	shift := byte(0)
	for i := uint32(0); i < b.TxCnt; i++ {
		tmp, ok := d.TXCache[b.TxArray[i]]
		tmpRes := false
		if !ok {
			tmp = new(CrossShardDec)
			tmpRes = b.Result(i)
			tmp.NewFromOther(b.ShardIndex, tmpRes)
			d.TXCache[b.TxArray[i]] = tmp
		} else {
			tmpRes = b.Result(i)
			tmp.UpdateFromOther(b.ShardIndex, tmpRes)
			if tmp.Total == 0 {
				d.UnlockTx(tmp.Data)
				delete(d.TXCache, b.TxArray[i])
			} else {
				d.TXCache[b.TxArray[i]] = tmp
			}
		}

		if b.ShardIndex == d.ShardNum {
			for j := uint32(0); j < b.MemCnt; j++ {
				tmp.Decision[shard.GlobalGroupMems[b.MemD[j].ID].InShardId] = (b.MemD[j].Decision[index]>>shift)&1 + 1
			}
			if tmpRes == false {
				for j := uint32(0); j < gVar.ShardSize; j++ {
					if tmp.Decision[j] == 1 {
						shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep += gVar.RepTN
					} else if tmp.Decision[j] == 2 {
						shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep -= gVar.RepFP
					}
				}
			}
		}
		if shift < 7 {
			shift++
		} else {
			index++
			shift = 0
		}
	}

	return nil
}

//GetTxBlock handle the txblock sent by the leader
func (d *DbRef) GetTxBlock(a *basic.TxBlock) error {

	for i := uint32(0); i < a.TxCnt; i++ {
		tmp, ok := d.TXCache[a.TxArray[i].Hash]
		if !ok {
			return fmt.Errorf("Verify txblock; No tx in cache")
		}
		if tmp.InCheckSum != 0 {
			return fmt.Errorf("Not be fully recognized")
		}
		for j := uint32(0); j < gVar.ShardSize; j++ {
			if tmp.Decision[j] == 1 {
				shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep -= gVar.RepFN
			} else if tmp.Decision[j] == 2 {
				shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep += gVar.RepTP
			}
		}
	}
	d.db.AddBlock(d.TxB)
	d.db.UpdateUTXO(a, d.ShardNum)
	return nil
}
