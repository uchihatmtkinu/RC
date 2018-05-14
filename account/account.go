package account

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"strconv"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/crypto"
)

//RcAcc the wallet of a user
type RcAcc struct {
	pri      ecdsa.PrivateKey
	Puk      ecdsa.PublicKey
	Addr     string
	AddrReal [32]byte
	AccType  int
	ID       string
}

//New generate a new wallet with different type, 0 is client and 1 is miner
func (acc *RcAcc) New(ID string, accType int) {
	h := sha256.New()
	h.Write([]byte(ID))
	var curve elliptic.Curve
	curve = elliptic.P256()
	var tmp *ecdsa.PrivateKey
	tmp, _ = ecdsa.GenerateKey(curve, rand.Reader)
	acc.pri = *tmp
	acc.Puk = acc.pri.PublicKey
	acc.AddrReal = cryptonew.AddressGenerate(&acc.pri)
	acc.Addr = base58.Encode(acc.AddrReal[:])
	acc.AccType = accType
}

//Load loading the account information
func (acc *RcAcc) Load(a1, a2, a3, a4, a5 string) {
	acc.pri.Curve = elliptic.P256()
	acc.pri.D = new(big.Int)
	acc.pri.D.SetString(a1, 10)
	acc.pri.X = new(big.Int)
	acc.pri.X.SetString(a2, 10)
	acc.pri.Y = new(big.Int)
	acc.pri.Y.SetString(a3, 10)
	acc.Puk = acc.pri.PublicKey
	acc.Addr = a4
	acc.AddrReal = cryptonew.AddressGenerate(&acc.pri)
	acc.AccType, _ = strconv.Atoi(a5)
}

//MakeTrans generate a transaction
func (acc *RcAcc) MakeTrans(In []basic.InType, Out []basic.OutType) basic.RawTransaction {
	var tmp basic.RawTransaction

	return tmp
}

//RetPri return the private key
func (acc *RcAcc) RetPri() ecdsa.PrivateKey {
	return acc.pri
}
