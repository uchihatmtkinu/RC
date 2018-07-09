package network

import (
	"time"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/rccache"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 16
const bufferSize = 1000
const timeoutCosi = 10 * time.Second //10seconds for timeout
const timeoutSync = 100 * time.Second
const timeSyncNotReadySleep = 5 * time.Second
const timeoutResponse = 120 * time.Second

//CurrentEpoch epoch now
var CurrentEpoch int

//LeaderAddr leader address
var LeaderAddr string

//MyGlobalID my global ID
var MyGlobalID int

//var AddrMapToInd map[string]int //ip+port
//var GroupMems []shard.MemShard
//GlobalAddrMapToInd
//var GlobalAddrMapToInd map[string]int

//CacheDbRef local database
var CacheDbRef rccache.DbRef

//------------------- shard process ----------------------
//readyInfo
type readyInfo struct {
	ID    int
	Epoch int
}

//readyMemberCh channel used in shard process, indicates the ready of the member for a new epoch
var readyMemberCh chan readyInfo
//readyLeaderCh channel used in shard process, indicates the ready of other shards for a new epoch
var readyLeaderCh chan readyInfo
//------------------- rep pow process -------------------------
//powInfo used in pow
type powInfo struct {
	ID    int
	Round int
	Hash  [32]byte
	Nonce int
}

//------------------- cosi process -------------------------
//commitInfo used in commitCh
type commitInfo struct {
	ID     int
	Commit cosi.Commitment
}

// challengeInfo challenge info
type challengeInfo struct {
	AggregatePublicKey ed25519.PublicKey
	AggregateCommit    cosi.Commitment
}

//responseInfo response info
type responseInfo struct {
	ID  int
	Sig cosi.SignaturePart
}

//channel used in cosi
//cosiAnnounceCh cosi announcement channel
var cosiAnnounceCh chan []byte

//cosiCommitCh		cosi commitment channel
var cosiCommitCh chan commitInfo
var cosiChallengeCh chan challengeInfo
var cosiResponseCh chan responseInfo
var cosiSigCh chan cosi.SignaturePart

//finalSignal
var finalSignal chan []byte

var startRep chan repInfo
var startTx chan int
var startSync chan bool

//syncSBInfo sync block info
type repInfo struct {
	Last bool
	Hash [][32]byte
}

//---------------------- sync process -------------
//syncSBInfo sync block info
type syncSBInfo struct {
	ID    int
	Block Reputation.SyncBlock
}

//syncTBInfo tx block info
type syncTBInfo struct {
	ID    int
	Block basic.TxBlock
}

//syncRequestInfo request sync
type syncRequestInfo struct {
	ID    int
	Epoch int
}

//TxBRequestInfo request txB
type TxBRequestInfo struct {
	Address string
	Shard   int
	Height  int
}

type syncNotReadyInfo struct {
	ID    int
	Epoch int
}

//channel used in sync
//syncCh
var syncSBCh [gVar.ShardCnt]chan syncSBInfo
var syncTBCh [gVar.ShardCnt]chan syncTBInfo
var syncNotReadyCh [gVar.ShardCnt]chan bool

//ShardDone flag determine whether the shard process is done
var ShardDone bool

//CoSiFlag flag determine the CoSi process has began
var CoSiFlag bool

//SyncFlag flag determine the Sync process has began
var SyncFlag bool

//IntialReadyCh channel used to indicate the process start
var IntialReadyCh chan bool
var ShardReadyCh chan bool
var CoSiReadyCh chan bool
var SyncReadyCh chan bool

var waitForFB chan bool

//FinalTxReadyCh whether the FB is done
var FinalTxReadyCh chan bool
