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
			if !d.DB.CheckUTXO(&a.In[i], a.Hash) {
				return false
			}
		}
	}
	return true
}

//LockTx locks all utxos related to transction a
func (d *DbRef) LockTx(a *basic.Transaction) error {
	d.DB.TXCache[a.Hash] = 1
	for i := uint32(0); i < a.TxinCnt; i++ {
		if a.In[i].ShardIndex() == d.ShardNum {
			err := d.DB.LockUTXO(&a.In[i])
			if err != nil {
				log.Panic(err)
			}
		}
	}
	return nil
}

//UnlockTx locks all utxos related to transction a
func (d *DbRef) UnlockTx(a *basic.Transaction) error {
	delete(d.DB.TXCache, a.Hash)
	for i := uint32(0); i < a.TxinCnt; i++ {
		if a.In[i].ShardIndex() == d.ShardNum {
			err := d.DB.UnlockUTXO(&a.In[i])
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
	}
	if tmp.InCheck[d.ShardNum] == 0 {
		if ok {
			delete(d.TXCache, a.Hash)
		}
		return fmt.Errorf("Not related TX")
	}
	d.TXCache[a.Hash] = tmp
	d.AddCache(a.Hash)
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
		tmpHash := b.TxArray[i]
		tmp, ok := d.TXCache[tmpHash]
		tmpRes := false
		if !ok {
			tmp = new(CrossShardDec)
			tmpRes = b.Result(i)
			tmp.NewFromOther(b.ShardIndex, tmpRes)
			d.TXCache[tmpHash] = tmp
		} else {
			tmpRes = b.Result(i)
			tmp.UpdateFromOther(b.ShardIndex, tmpRes)
			if tmp.Total == 0 { //Review
				d.UnlockTx(tmp.Data)
				delete(d.TXCache, tmpHash)
			} else {
				d.TXCache[tmpHash] = tmp
			}
		}

		if b.ShardIndex == d.ShardNum {
			for j := uint32(0); j < b.MemCnt; j++ {
				tmp.Decision[shard.GlobalGroupMems[b.MemD[j].ID].InShardId] = (b.MemD[j].Decision[index]>>shift)&1 + 1
			}
			if tmpRes == false {
				for j := uint32(0); j < gVar.ShardSize; j++ {
					if tmp.Decision[j] == 1 {
						shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep += gVar.RepTN * int64(tmp.Value)
					} else if tmp.Decision[j] == 2 {
						shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep -= gVar.RepFP * int64(tmp.Value)
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
	if a.Kind != 1 {
		return fmt.Errorf("Not valid txblock type")
	}
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
				shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep -= gVar.RepFN * int64(tmp.Value)
			} else if tmp.Decision[j] == 2 {
				shard.GlobalGroupMems[shard.ShardToGlobal[d.ShardNum][j]].Rep += gVar.RepTP * int64(tmp.Value)
			}
		}
		d.ClearCache(a.TxArray[i].Hash)
	}
	d.DB.AddBlock(a)
	d.DB.UpdateUTXO(a, d.ShardNum)
	return nil
}

//GetFinalTxBlock handle the txblock sent by the leader
func (d *DbRef) GetFinalTxBlock(a *basic.TxBlock) error {
	if a.ShardID == d.ShardNum {

	} else {

	}
	return nil
}
