package account

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"math/big"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/cryptonew"
	"github.com/uchihatmtkinu/RC/ed25519"
)

//RcAcc the wallet of a user1
type RcAcc struct {
	pri      ecdsa.PrivateKey
	Puk      ecdsa.PublicKey
	cosiPri	 ed25519.PrivateKey
	CosiPuk	 ed25519.PublicKey
	Addr     string
	AddrReal [32]byte
	//AccType  int
	ID       string
	Rep		 int //total reputation among a period of time
}

//New generate a new wallet with different type
func (acc *RcAcc) New(ID string) {
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
	//acc.AccType = accType
}

func (acc *RcAcc) NewCosi() {
	pubKey, priKey, _ := ed25519.GenerateKey(nil)
	acc.cosiPri = priKey
	acc.CosiPuk = pubKey
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
	//acc.AccType, _ = strconv.Atoi(a5)
}

//MakeTrans generate a transaction
func (acc *RcAcc) MakeTrans(In []basic.InType, Out []basic.OutType) basic.Transaction {
	var tmp basic.Transaction

	return tmp
}

//RetPri return the private key
func (acc *RcAcc) RetPri() ecdsa.PrivateKey {
	return acc.pri
}
