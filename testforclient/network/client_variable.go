package network

import (
	"time"

	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/rccache"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12
const bufferSize = 1000
const timeoutCosi = 10 * time.Second //10seconds for timeout

var LeaderAddr string

//var AddrMapToInd map[string]int //ip+port
//var GroupMems []shard.MemShard

var GlobalAddrMapToInd map[string]int


var CacheDbRef rccache.DbRef

//used in commitCh
type commitInfoCh struct {
	addr    string
	request []byte
}

type challengeMessage struct {
	aggregatePublicKey ed25519.PublicKey
	aggregateCommit    cosi.Commitment
}

type responseInfoCh struct {
	addr    string
	request []byte
}

type cosiSigMessage struct {
	pubKeys []ed25519.PublicKey
	cosiSig cosi.SignaturePart
}

//channel used in cosi
var cosiAnnounceCh chan []byte
var cosiCommitCh chan commitInfoCh
var cosiChallengeCh chan []byte
var cosiResponseCh chan responseInfoCh
var cosiSigCh chan []byte

//channel used in pow
var repPowTxCh chan []byte
var repPowRxCh chan []byte
