package network

import (
	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/account"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/shard"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12
const bufferSize = 1000
const timeoutCosi = 10 //10seconds for timeout

var MyAccount account.RcAcc

var LeaderAddr string
var AddrMapToInd map[string]int //ip+port
var GroupMems []shard.MemShard
var NumMems int

//used in rep calculation, scaling factor
const RepTP = 1
const RepTN = 1
const RepFP = 1
const RepFN = 1

//used in pow
var MyPoW Reputation.ProofOfWork

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

//channel
var cosiAnnounceCh chan []byte
var cosiCommitCh chan commitInfoCh
var cosiChallengeCh chan []byte
var cosiResponseCh chan responseInfoCh
var cosiSigCh chan []byte