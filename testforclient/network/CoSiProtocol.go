package network


import (
	"fmt"
	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"github.com/uchihatmtkinu/RC/shard"
	"time"
	"bytes"
	"log"
	"encoding/gob"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/account"
)


// myCommit my cosi commitment
var myCommit 	cosi.Commitment
var mySecret 	*cosi.Secret
var cosimask	[]byte
var commits		[]cosi.Commitment
var	pubKeys		[]ed25519.PublicKey
var sigParts 	[]cosi.SignaturePart
var sbMessage	[]byte
var cosiSig		cosi.SignaturePart


// leaderCosiProcess leader use this
func leaderCosiProcess(ms *[]shard.MemShard, prevRepBlockHash [32]byte) cosi.SignaturePart{
	//initialize
	var sbMessage []byte
	var it *shard.MemShard
	sbMessage = prevRepBlockHash[:]

	commits = make([]cosi.Commitment, shard.NumMems)
	pubKeys = make([]ed25519.PublicKey, shard.NumMems)
	myCommit, mySecret, _ = cosi.Commit(nil)

	//byte mask 0-7 bit in one byte represent user 0-7, 8-15...
	intilizeMaskBit(&cosimask, (shard.NumMems+7)>>3,false)
	//announcement
	for i:=0; i < shard.NumMems; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		pubKeys[it.InShardId] = it.CosiPub
		sendCosiMessage(it.Address, "cosiAnnoun", sbMessage)
	}
	//handle commits
	timeoutflag := true
	for timeoutflag {
		select {
		case commitInfo := <-cosiCommitCh:
			commits[(*ms)[GlobalAddrMapToInd[commitInfo.addr]].InShardId] = handleCommit(commitInfo.request)
			setMaskBit((*ms)[GlobalAddrMapToInd[commitInfo.addr]].InShardId, cosi.Enabled, &cosimask)
		case <-time.After(10 * time.Second):
			timeoutflag = false
		}
	}
	//handle leader's commit
	commits[(*ms)[GlobalAddrMapToInd[account.MyAccount.Addr]].InShardId] = myCommit
	setMaskBit((*ms)[GlobalAddrMapToInd[account.MyAccount.Addr]].InShardId, cosi.Enabled, &cosimask)
	close(cosiCommitCh)

	// The leader then combines these into an aggregate commit.
	cosigners := cosi.NewCosigners(pubKeys, cosimask)
	aggregatePublicKey := cosigners.AggregatePublicKey()
	aggregateCommit := cosigners.AggregateCommit(commits)
	currentChaMessage := challengeMessage{aggregatePublicKey, aggregateCommit}

	//sign or challenge
	for i:=uint32(0); i <gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		if maskBit(it.InShardId, &cosimask)==cosi.Enabled {
			sendCosiMessage(it.Address, "cosiChallen", currentChaMessage)
		}
	}
	//handle response
	sigParts = make([]cosi.SignaturePart, shard.NumMems)
	timeoutflag = true
	for timeoutflag {
		select {
		case reponseInfo := <-cosiResponseCh:
			sigParts[(*ms)[GlobalAddrMapToInd[reponseInfo.addr]].InShardId] = handleResponse(reponseInfo.request)
		case <-time.After(10 * time.Second):
			timeoutflag = false
		}
	}

	mySigPart := cosi.Cosign(account.MyAccount.CosiPri, mySecret, sbMessage, aggregatePublicKey, aggregateCommit)
	sigParts[(*ms)[GlobalAddrMapToInd[account.MyAccount.Addr]].InShardId] = mySigPart
	close(cosiResponseCh)
	// Finally, the leader combines the two signature parts
	// into a final collective signature.
	cosiSig = cosigners.AggregateSignature(aggregateCommit, sigParts)
	currentSigMessage := cosiSigMessage{pubKeys,cosiSig}
	for i:=uint32(0); i <gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		if maskBit(it.InShardId, &cosimask)==cosi.Enabled {
			sendCosiMessage(it.Address, "cosiSig", currentSigMessage)
		}
	}

	return cosiSig
}

// memberCosiProcess member use this
func memberCosiProcess(prevRepBlockHash [32]byte) (bool, []byte){
	var sbMessage []byte
	sbMessage = prevRepBlockHash[:]
	leaderSBMessage := <-cosiAnnounceCh
	if !verifySBMessage(sbMessage, handleAnnounce(leaderSBMessage)) {
		fmt.Println("Sync Block from leader is wrong!")
		//send warning
	}
	myCommit, mySecret, _ = cosi.Commit(nil)
	sendCosiMessage(LeaderAddr, "cosiCommit", myCommit)
	request := <- cosiChallengeCh
	currentChaMessage := handleChallenge(request)
	sigPart := cosi.Cosign(account.MyAccount.CosiPri, mySecret, sbMessage, currentChaMessage.aggregatePublicKey, currentChaMessage.aggregateCommit)
	sendCosiMessage(LeaderAddr, "cosiRespon", sigPart)
	request = <- cosiSigCh
	currentSigMessage := handleCosiSig(request) // handleCosiSig is the same as handleResponse
	valid := cosi.Verify(currentSigMessage.pubKeys, nil, sbMessage, currentSigMessage.cosiSig)
	return valid, currentSigMessage.cosiSig[64:]
}

// handleCommit rx commit
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


// handleAnnounce rx announce
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

// handleChallenge rx challenge
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

// handleResponse rx response
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

// handleCosiSig rx cosisig
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

// sendCosiMessage send cosi message
func sendCosiMessage(addr string, command string, message interface{}, ) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}

// verifySBMessage compare whether the message from leader is the same as itself
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
