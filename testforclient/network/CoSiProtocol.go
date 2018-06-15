package network


import (
	"fmt"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/Reputation"
	"time"
	"bytes"
	"log"
	"encoding/gob"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/account"
)

var timeoutflag bool
var myCommit 	cosi.Commitment
var mySecret 	*cosi.Secret
var cosimask	[]byte
var commits		[]cosi.Commitment
var	pubKeys		[]ed25519.PublicKey
var sigParts 	[]cosi.SignaturePart
var sbMessage	[]byte
var cosiSig		cosi.SignaturePart


//leader use this
func leaderCosiProcess(ms *[]shard.MemShard, sb *Reputation.SyncBlock) cosi.SignaturePart{
	//initialize
	var it *shard.MemShard
	sbMessage := sb.Hash
	cosimask = make([]byte, (shard.NumMems+7)>>3) //byte mask 0-7 bit in one byte represent user 0-7, 8-15...
	commits = make([]cosi.Commitment, shard.NumMems)
	pubKeys = make([]ed25519.PublicKey, shard.NumMems)
	myCommit, mySecret, _ = cosi.Commit(nil)
	for i := range cosimask {
		cosimask[i] = 0xff // all disabled
	}

	//announcement
	for i:=uint32(0); i <gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		pubKeys[it.InShardId] = it.CosiPub
		sendCosiMessage(it.Address, "cosiAnnouncement", sbMessage)
	}
	//handle commits
	timeoutflag = true
	for timeoutflag {
		select {
		case commitInfo := <-cosiCommitCh:
			commits[GlobalAddrMapToInd[commitInfo.addr]] = handleCommit(commitInfo.request)
			setMaskBit(GlobalAddrMapToInd[commitInfo.addr], cosi.Enabled)
		case <-time.After(10 * time.Second):
			timeoutflag = false
		}
	}
	//handle leader's commit
	commits[GlobalAddrMapToInd[account.MyAccount.Addr]] = myCommit
	setMaskBit(GlobalAddrMapToInd[account.MyAccount.Addr], cosi.Enabled)
	close(cosiCommitCh)

	// The leader then combines these into an aggregate commit.
	cosigners := cosi.NewCosigners(pubKeys, cosimask)
	aggregatePublicKey := cosigners.AggregatePublicKey()
	aggregateCommit := cosigners.AggregateCommit(commits)
	currentChaMessage := challengeMessage{aggregatePublicKey, aggregateCommit}

	//sign or challenge
	for i:=uint32(0); i <gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		if maskBit(it.InShardId)==cosi.Enabled {
			sendCosiMessage(it.Address, "cosiChallenge", currentChaMessage)
		}
	}
	//handle response
	sigParts = make([]cosi.SignaturePart, shard.NumMems)
	timeoutflag = true
	for timeoutflag {
		select {
		case reponseInfo := <-cosiResponseCh:
			sigParts[GlobalAddrMapToInd[reponseInfo.addr]] = handleResponse(reponseInfo.request)
		case <-time.After(10 * time.Second):
			timeoutflag = false
		}
	}

	mySigPart := cosi.Cosign(account.MyAccount.CosiPri, mySecret, sbMessage, aggregatePublicKey, aggregateCommit)
	sigParts[GlobalAddrMapToInd[account.MyAccount.Addr]] = mySigPart

	// Finally, the leader combines the two signature parts
	// into a final collective signature.
	cosiSig = cosigners.AggregateSignature(aggregateCommit, sigParts)
	currentSigMessage := cosiSigMessage{pubKeys,cosiSig}
	for i:=uint32(0); i <gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		if maskBit(it.InShardId)==cosi.Enabled {
			sendCosiMessage(it.Address, "cosiSig", currentSigMessage)
		}
	}

	return cosiSig
}

//member use this
func memberCosiProcess(sb *Reputation.SyncBlock) (bool){
	sbMessage = sb.Hash
	leaderSBMessage := <-cosiAnnounceCh
	if !verifySBMessage(sbMessage, handleAnnounce(leaderSBMessage)) {
		fmt.Println("Sync Block from leader is wrong!")
	}
	myCommit, mySecret, _ = cosi.Commit(nil)
	sendCosiMessage(LeaderAddr, "cosicommit", myCommit)
	request := <- cosiChallengeCh
	currentChaMessage := handleChallenge(request)
	sigPart := cosi.Cosign(account.MyAccount.CosiPri, mySecret, sbMessage, currentChaMessage.aggregatePublicKey, currentChaMessage.aggregateCommit)
	sendCosiMessage(LeaderAddr, "cosiresponse", sigPart)
	request = <- cosiSigCh
	currentSigMessage := handleCosiSig(request) // handleCosiSig is the same as handleResponse
	valid := cosi.Verify(currentSigMessage.pubKeys, nil, sbMessage, currentSigMessage.cosiSig)
	return valid
}

func handleCommit(request []byte) []byte{
	var buff bytes.Buffer
	var payload []byte

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	return payload
}



func handleAnnounce(request []byte) []byte {
	var buff bytes.Buffer
	var payload []byte

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	return payload
}

func handleChallenge(request []byte) challengeMessage {
	var buff bytes.Buffer
	var payload challengeMessage

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	return payload
}

func handleResponse(request[]byte) cosi.SignaturePart{
	var buff bytes.Buffer
	var payload cosi.SignaturePart

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	return payload
}

func handleCosiSig(request []byte) cosiSigMessage {
	var buff bytes.Buffer
	var payload cosiSigMessage

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	return payload
}

func sendCosiMessage(addr string, command string, message interface{}, ) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

//compare whether the message from leader is the same as itself
func verifySBMessage(a,b []byte) bool{
	if a == nil && b == nil {
		return true;
	}
	if a == nil || b == nil {
		return false;
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// enable = 0 = false, disable = 1= true
func setMaskBit(signer int, value cosi.MaskBit) {
	byt := signer >> 3

	bit := byte(1) << uint(signer&7)
	if value == cosi.Disabled { // disable
		if cosimask[byt]&bit == 0 { // was enabled
			cosimask[byt] |= bit // disable it
		}
	} else { // enable
		if cosimask[byt]&bit != 0 { // was disabled
			cosimask[byt] &^= bit
		}
	}
}

// MaskBit returns a boolean value indicating whether
// the indicated signer is Enabled or Disabled.
func maskBit(signer int) (value cosi.MaskBit) {
	byt := signer >> 3
	bit := byte(1) << uint(signer&7)
	return (cosimask[byt] & bit) != 0
}
