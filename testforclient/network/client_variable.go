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

//currentEpoch epoch now
var CurrentEpoch int
//LeaderAddr leader address
var LeaderAddr string
//MyGlobalID my global ID
var MyGlobalID int
//var AddrMapToInd map[string]int //ip+port
//var GroupMems []shard.MemShard
//GlobalAddrMapToInd
var GlobalAddrMapToInd map[string]int

var CacheDbRef rccache.DbRef



//used in commitCh
type commitInfoCh struct {
	Addr    string
	Commit  cosi.Commitment
}

// challenge info
type challengeMessage struct {
	AggregatePublicKey ed25519.PublicKey
	AggregateCommit    cosi.Commitment
}

//response info
type responseInfoCh struct {
	Addr    string
	Sig cosi.SignaturePart
}


//channel used in cosi
//cosiAnnounceCh cosi announcement channel
var cosiAnnounceCh 	chan []byte
//cosiCommitCh		cosi commitment channel
var cosiCommitCh 	chan commitInfoCh
var cosiChallengeCh chan challengeMessage
var cosiResponseCh 	chan responseInfoCh
var cosiSigCh  		chan cosi.SignaturePart




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


var IntialReadyCh chan bool