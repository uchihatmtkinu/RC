package network

import (
	"time"

	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/rccache"
	"sync"
	"github.com/uchihatmtkinu/RC/gVar"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12
const bufferSize = 1000
const timeoutCosi = 10 * time.Second //10seconds for timeout
const timeoutSync = 20 * time.Second
//LeaderAddr leader address
var LeaderAddr string

//var AddrMapToInd map[string]int //ip+port
//var GroupMems []shard.MemShard
//GlobalAddrMapToInd
var GlobalAddrMapToInd map[string]int

var CacheDbRef rccache.DbRef

//SafeCounter used in
type safeCounter struct {
	cnt	int
	mux sync.Mutex
}


//used in commitCh
type commitInfoCh struct {
	addr    string
	request []byte
}

// challenge info
type challengeMessage struct {
	aggregatePublicKey ed25519.PublicKey
	aggregateCommit    cosi.Commitment
}

//response info
type responseInfoCh struct {
	addr    string
	request []byte
}

//cosisig info
type cosiSigMessage struct {
	pubKeys []ed25519.PublicKey
	cosiSig cosi.SignaturePart
}

//channel used in cosi
//cosiAnnounceCh cosi announcement channel
var cosiAnnounceCh chan []byte
//cosiCommitCh	cosi commitment channel
var cosiCommitCh chan commitInfoCh
var cosiChallengeCh chan []byte
var cosiResponseCh chan responseInfoCh
var cosiSigCh chan []byte

//channel used in pow
//repPowTxCh chan []byte
//repPowRxCh reppow receive channel
var repPowRxCh chan []byte

//sbInfoCh
type sbInfoCh struct {
	id		int
	request	[]byte
}

//channel used in sync
//syncCh
var syncSBCh [gVar.ShardCnt] chan sbInfoCh
var syncTBCh [gVar.ShardCnt] chan sbInfoCh
var sbRxFlag [gVar.ShardCnt] chan bool
var tbRxFlag [gVar.ShardCnt] chan bool