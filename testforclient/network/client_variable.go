package network

import (
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/rccache"
	"github.com/uchihatmtkinu/RC/gVar"
	"sync"
	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/basic"
	"time"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 16
const bufferSize = 1000
const timeoutCosi = 10 * time.Second //10seconds for timeout
const timeoutSync = 20 * time.Second
const timeSyncNotReadySleep = 5 * time.Second
const timeoutResponse = 120 * time.Second
//LeaderAddr leader address
var LeaderAddr string

//var AddrMapToInd map[string]int //ip+port
//var GroupMems []shard.MemShard
//GlobalAddrMapToInd
var GlobalAddrMapToInd map[string]int

var CacheDbRef rccache.DbRef



//used in commitCh
type commitInfoCh struct {
	addr    string
	commit []byte
}

// challenge info
type challengeMessage struct {
	aggregatePublicKey ed25519.PublicKey
	aggregateCommit    cosi.Commitment
}

//response info
type responseInfoCh struct {
	addr    string
	sig cosi.SignaturePart
}

//cosisig info
type cosiSigMessage struct {
	pubKeys []ed25519.PublicKey
	cosiSig cosi.SignaturePart
}

//channel used in cosi
//cosiAnnounceCh cosi announcement channel
var cosiAnnounceCh 	chan []byte
//cosiCommitCh		cosi commitment channel
var cosiCommitCh 	chan commitInfoCh
var cosiChallengeCh chan challengeMessage
var cosiResponseCh 	chan responseInfoCh
var cosiSigCh  		chan cosi.SignaturePart



//sbInfoCh
type sbInfoCh struct {
	id		int
	block 	Reputation.SyncBlock
}

//tbInfoCh
type tbInfoCh struct {
	id		int
	block 	basic.TxBlock
}

//channel used in sync
//syncCh
var syncSBCh [gVar.ShardCnt] 		chan Reputation.SyncBlock
var syncTBCh [gVar.ShardCnt]	 	chan basic.TxBlock
var syncNotReadyCh [gVar.ShardCnt]	chan bool



//safeCounter used in
type safeCounter struct {
	cnt	int
	mux sync.Mutex
}



//readyCh channel for ready a new epoch
var readyCh	chan string


//CoSiFlag flag determine the process has began
var CoSiFlag	bool

//syncReady sync is ready
var syncReady	bool