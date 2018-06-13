package rccache

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/uchihatmtkinu/RC/basic"
)

//VerifyTx verify the utxos related to transaction a
func (d *dbRef) VerifyTx(a *basic.Transaction) bool {
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
func (d *dbRef) LockTx(a *basic.Transaction) error {
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
func (d *dbRef) UnlockTx(a *basic.Transaction) error {
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

//GetTxList and store those transactions
func (d *dbRef) GetTxList(a *basic.TxList) error {
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp, ok := d.TXCache[a.TxArray[i].Hash]
		if ok {
			tmp.Update(&a.TxArray[i])
		} else {
			tmp = new(CrossShardDec)
			tmp.New(&a.TxArray[i])
			d.TXCache[a.TxArray[i].Hash] = tmp
		}
	}
	return nil
}

//ProcessTL verify the transactions in the txlist
func (d *dbRef) ProcessTL(a *basic.TxList) error {
	d.TLNow = new(basic.TxDecision)
	d.TLNow.Set(d.ID, d.ShardNum, 0)
	d.TLNow.HashID = a.HashID
	d.TLNow.Single = 0
	d.TLNow.Sig = make([]basic.RCSign, 0, basic.ShardCnt)
	var tmpHash [basic.ShardCnt][]byte
	var tmpDecision [basic.ShardCnt]basic.TxDecision
	for i := uint32(0); i < basic.ShardCnt; i++ {
		tmpHash[i] = []byte{}
		tmpDecision[i].Set(d.ID, i, 1)
	}
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp, _ := d.TXCache[a.TxArray[i].Hash]
		var res byte
		if tmp.InCheck[d.ShardNum] == 3 {
			if d.VerifyTx(&a.TxArray[i]) {
				res = byte(1)
				d.LockTx(&a.TxArray[i])
			}
			d.TLNow.Add(res)
			for j := 0; j < len(tmp.ShardRelated); j++ {
				tmpHash[tmp.ShardRelated[j]] = append(tmpHash[tmp.ShardRelated[j]], a.TxArray[i].Hash[:]...)
				tmpDecision[tmp.ShardRelated[j]].Add(res)
			}
		}
	}
	for i := uint32(0); i < basic.ShardCnt; i++ {
		tmpDecision[i].HashID = sha256.Sum256(tmpHash[i])
		tmpDecision[i].Sign(&d.prk, 0)
		d.TLNow.Sig[i] = tmpDecision[i].Sig[0]
	}
	d.TLSent = d.TLNow
	return nil
}

//GetTxDecSet and ready to verify txblock
func (d *dbRef) GetTDS(b *basic.TxDecSet) error {
	for i := uint32(0); i < b.TxCnt; i++ {
		tmp, ok := d.TXCache[b.TxArray[i]]
		if !ok {
			tmp = new(CrossShardDec)
			tmp.NewFromOther(b.ShardIndex, b.Result(i))
			d.TXCache[b.TxArray[i]] = tmp
		} else {
			tmp.UpdateFromOther(b.ShardIndex, b.Result(i))
			if tmp.Total == 0 {
				d.UnlockTx(tmp.Data)
				delete(d.TXCache, b.TxArray[i])
			} else {
				d.TXCache[b.TxArray[i]] = tmp
			}
		}
	}
	return nil
}

//GetTxBlock handle the txblock sent by the leader
func (d *dbRef) GetTxBlock(a *basic.TxBlock) error {
	for i := uint32(0); i < a.TxCnt; i++ {
		tmp, ok := d.TXCache[a.TxArray[i].Hash]
		if !ok {
			return fmt.Errorf("Verify txblock; No tx in cache")
		}
		if tmp.InCheckSum != 0 {
			return fmt.Errorf("Not be fully recognized")
		}
	}
	d.db.AddBlock(d.TxB)
	d.db.UpdateUTXO(a, d.ShardNum)
	return nil
}
