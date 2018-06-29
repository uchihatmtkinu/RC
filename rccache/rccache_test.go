package rccache

import (
	"crypto/x509"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/uchihatmtkinu/RC/basic"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/cryptonew"
	"github.com/uchihatmtkinu/RC/gVar"

	"github.com/uchihatmtkinu/RC/shard"

	"github.com/uchihatmtkinu/RC/account"
)

func TestGeneratePriKey(t *testing.T) {
	file, _ := os.Create("PriKeys.txt")
	for i := 0; i < 1800; i++ {
		var tmp account.RcAcc
		tmp.New(strconv.Itoa(i))
		fmt.Println(tmp.Pri)
		tmpHash, _ := x509.MarshalECPrivateKey(&tmp.Pri)
		file.Write(tmpHash)
		file.Write(tmp.CosiPri)
	}
	file.Close()
}

func TestOutToData(t *testing.T) {
	numCnt := 4
	acc := make([]account.RcAcc, numCnt)
	dbs := make([]DbRef, numCnt)
	shard.GlobalGroupMems = make([]shard.MemShard, numCnt)
	file, ok := os.Open("PriKeys.txt")
	if ok != nil {
		t.Error("No file")
	}
	accWallet := make([]basic.AccCache, numCnt)
	for i := 0; i < numCnt; i++ {
		acc[i].New(strconv.Itoa(i))
		acc[i].NewCosi()
		tmp1 := make([]byte, 121)
		tmp2 := make([]byte, 64)
		file.Read(tmp1)
		xxx, _ := x509.ParseECPrivateKey(tmp1)
		acc[i].Pri = *xxx
		acc[i].Puk = acc[i].Pri.PublicKey
		acc[i].CosiPri = tmp2
		acc[i].CosiPuk = tmp2[32:]
		acc[i].AddrReal = cryptonew.AddressGenerate(&acc[i].Pri)
		acc[i].Addr = base58.Encode(acc[i].AddrReal[:])
		accWallet[i].ID = acc[i].AddrReal
		accWallet[i].Value = 100
		//tmp, _ := x509.MarshalECPrivateKey(&acc[i].Pri)
<<<<<<< HEAD
		shard.GlobalGroupMems[i].NewMemShard(&acc[i], "123")
=======
		shard.GlobalGroupMems[i].NewMemShard(&acc[i],"123")
>>>>>>> 5516540c6d8d0a298291f239ca9841087b081fdc
		dbs[i].New(uint32(i), acc[i].Pri)
	}
	t.Error("Check1")
	for i := 0; i < numCnt; i++ {
		for j := 0; j < numCnt; j++ {
			//dbs[i].DB.AddAccount(&accWallet[j])
		}
		for j := 0; j < int(gVar.ShardCnt); j++ {
			fmt.Println(i+'-', j, ": ", dbs[i].DB.LastFB[j])
		}
		//dbs[i].DB.ShowAccount()
	}
	shard.ShardToGlobal = make([][]int, gVar.ShardCnt)
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		shard.ShardToGlobal[i] = make([]int, gVar.ShardSize)
		for j := uint32(0); j < gVar.ShardSize; j++ {
			shard.ShardToGlobal[i][j] = int(i*2 + j)
		}
	}
	file.Close()
	t.Error("Check")
}
