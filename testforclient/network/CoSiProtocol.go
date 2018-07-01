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
	"github.com/uchihatmtkinu/RC/Reputation"
)


// leaderCosiProcess leader use this
func LeaderCosiProcess(ms *[]shard.MemShard, prevRepBlockHash [32]byte) cosi.SignaturePart{
	//initialize
	// myCommit my cosi commitment
	var myCommit 	cosi.Commitment
	var mySecret 	*cosi.Secret
	var sbMessage []byte
	var it *shard.MemShard
	var cosimask	[]byte
	var responsemask[]byte
	var commits		[]cosi.Commitment
	var	pubKeys		[]ed25519.PublicKey
	var sigParts 	[]cosi.SignaturePart
	var cosiSig		cosi.SignaturePart
	var signMemNum	int
	//clear ready channel for shardProcess
	for len(readyCh) > 0 {
		<-readyCh
	}
	// cosi begin
	fmt.Println("Leader CoSi")
	CoSiFlag = true
	//To simplify the problem, we just validate the previous repblock hash
	sbMessage = prevRepBlockHash[:]
	commits = make([]cosi.Commitment, shard.NumMems)
	pubKeys = make([]ed25519.PublicKey, shard.NumMems)

	myCommit, mySecret, _ = cosi.Commit(nil)

	//byte mask 0-7 bit in one byte represent user 0-7, 8-15...
	intilizeMaskBit(&cosimask, (shard.NumMems+7)>>3,cosi.Disabled)
	intilizeMaskBit(&responsemask, (shard.NumMems+7)>>3,cosi.Disabled)
	//announcement
	cosiCommitCh = make(chan commitInfoCh, bufferSize)

	//handle leader's commit
	commits[(*ms)[GlobalAddrMapToInd[shard.MyMenShard.Address]].InShardId] = myCommit
	setMaskBit((*ms)[GlobalAddrMapToInd[shard.MyMenShard.Address]].InShardId, cosi.Enabled, &cosimask)
	setMaskBit((*ms)[GlobalAddrMapToInd[shard.MyMenShard.Address]].InShardId, cosi.Enabled, &responsemask)


	for i:=1; i < int(gVar.ShardSize); i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		pubKeys[it.InShardId] = it.CosiPub
		if i!=0 {
			SendCosiMessage(it.Address, "cosiAnnoun", sbMessage)
		}
	}
	fmt.Println("sent CoSi announce")
	//handle members' commits
	signMemNum = 1
	timeoutflag := true
	annoucementFlag := true
	for timeoutflag && annoucementFlag{
		select {
		case commitInfo := <-cosiCommitCh:
			commits[(*ms)[GlobalAddrMapToInd[commitInfo.Addr]].InShardId] = commitInfo.Commit
			fmt.Println("IP",commitInfo.Addr)
			fmt.Println("global ID",GlobalAddrMapToInd[commitInfo.Addr])
			setMaskBit((*ms)[GlobalAddrMapToInd[commitInfo.Addr]].InShardId, cosi.Enabled, &cosimask)
			signMemNum++
			if signMemNum == int(gVar.ShardSize){
				annoucementFlag = false
			}
			//fmt.Println("received commit from",(*ms)[GlobalAddrMapToInd[commitInfo.addr]].InShardId)
			//timeoutflag = false
		case <-time.After(timeoutCosi):
			//resend after 20 seconds
			for i:=uint32(1); i <gVar.ShardSize; i++ {
				it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
				if maskBit(it.InShardId, &cosimask)==cosi.Disabled {
					SendCosiMessage(it.Address, "cosiAnnoun", sbMessage)
				}
			}
		case <-time.After(timeoutResponse):
			timeoutflag = false
		}
	}
	fmt.Println("Recived CoSi comit")
	//fmt.Println((*ms)[GlobalAddrMapToInd[shard.MyMenShard.Address]].InShardId)

	// The leader then combines these into an aggregate commitment.
	cosigners := cosi.NewCosigners(pubKeys, cosimask)
	aggregatePublicKey := cosigners.AggregatePublicKey()
	aggregateCommit := cosigners.AggregateCommit(commits[:])
	currentChaMessage := challengeMessage{aggregatePublicKey, aggregateCommit}

	//sign or challenge
	cosiResponseCh = make(chan responseInfoCh, bufferSize)
	for i:=uint32(1); i <gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		if maskBit(it.InShardId, &cosimask)==cosi.Enabled {
			SendCosiMessage(it.Address, "cosiChallen", currentChaMessage)
		}
	}
	fmt.Println("Sent CoSi Challenage")
	//handle response
	sigParts = make([]cosi.SignaturePart, shard.NumMems)
	sigCount := 1
	//timeoutflag = true
	for sigCount < signMemNum{
		select {
		case reponseInfo := <-cosiResponseCh:
			it = &(*ms)[GlobalAddrMapToInd[reponseInfo.Addr]]
			sigParts[it.InShardId] = reponseInfo.Sig
			setMaskBit(it.InShardId, cosi.Disabled, &responsemask)
			sigCount++
			fmt.Println("received response")
		case <-time.After(timeoutCosi):
			//resend after 20 seconds
			for i:=uint32(1); i <gVar.ShardSize; i++ {
				it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
				if maskBit(it.InShardId, &responsemask)==cosi.Enabled {
					SendCosiMessage(it.Address, "cosiChallen", currentChaMessage)
				}
			}
		//case <- time.After(timeoutResponse):
		//	timeoutflag = false
		}
	}
	fmt.Println("Received CoSi Response")
	mySigPart := cosi.Cosign(shard.MyMenShard.RealAccount.CosiPri, mySecret, sbMessage, aggregatePublicKey, aggregateCommit)
	sigParts[(*ms)[GlobalAddrMapToInd[shard.MyMenShard.Address]].InShardId] = mySigPart

	// Finally, the leader combines the two signature parts
	// into a final collective signature.
	cosiSig = cosigners.AggregateSignature(aggregateCommit, sigParts)
	//currentSigMessage := cosiSigMessage{pubKeys,cosiSig}
	for i:=uint32(1); i <gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		if maskBit(it.InShardId, &cosimask)==cosi.Enabled {
			SendCosiMessage(it.Address, "cosiSig", cosiSig)
		}
	}

	//Add sync block
	Reputation.MyRepBlockChain.AddSyncBlock(ms, cosiSig)
	//close CoSi
	CoSiFlag =false
	close(cosiCommitCh)
	close(cosiResponseCh)
	fmt.Println("CoSi finished, result is ", cosiSig)
	return cosiSig
}

// MemberCosiProcess member use this
func MemberCosiProcess(ms *[]shard.MemShard, prevRepBlockHash [32]byte) (bool, []byte){
	var sbMessage []byte
	// myCommit my cosi commitment
	var myCommit 	cosi.Commitment
	var mySecret 	*cosi.Secret
	var	pubKeys		[]ed25519.PublicKey
	var it *shard.MemShard
	//var timeoutflag bool
	//timeoutflag = false
	cosiAnnounceCh = make(chan []byte)
	cosiChallengeCh = make(chan challengeMessage)
	cosiSigCh = make(chan cosi.SignaturePart)
	fmt.Println("Member CoSi")
	//generate pubKeys
	pubKeys = make([]ed25519.PublicKey, shard.NumMems)
	for i:=0; i < shard.NumMems; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		pubKeys[it.InShardId] = it.CosiPub
	}

	//receive announce and verify message
	//TODO add timeout
	sbMessage = prevRepBlockHash[:]
	leaderSBMessage := <-cosiAnnounceCh
	//close(cosiAnnounceCh)
	if !verifySBMessage(sbMessage, leaderSBMessage) {
		fmt.Println("Sync Block from leader is wrong!")
		//TODO send warning
	}
	fmt.Println("received cosi announce")
	//send commit

	myCommit, mySecret, _ = cosi.Commit(nil)
	SendCosiMessage(LeaderAddr, "cosiCommit", myCommit)
	fmt.Println("sent cosi commit")
	//receive challenge
	currentChaMessage :=  <- cosiChallengeCh

	fmt.Println("received cosi challenge from leader")
	//send signature

	sigPart := cosi.Cosign(shard.MyMenShard.RealAccount.CosiPri, mySecret, sbMessage, currentChaMessage.AggregatePublicKey, currentChaMessage.AggregateCommit)
	SendCosiMessage(LeaderAddr, "cosiRespon", sigPart)

	//receive cosisig and verify
	cosiSigMessage := <- cosiSigCh

	valid := cosi.Verify(pubKeys, nil, sbMessage, cosiSigMessage)
	//add sync block
	if valid {
		Reputation.MyRepBlockChain.AddSyncBlock(ms, cosiSigMessage)
	}
	//close cosi
	CoSiFlag =false
	close(cosiChallengeCh)
	close(cosiSigCh)
	fmt.Println("Member CoSi finished, result is ", valid)
	return valid, cosiSigMessage
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

// SendCosiMessage send cosi message
func SendCosiMessage(addr string, command string, message interface{}, ) {
	payload := gobEncode(message)
	request := append(commandToBytes(command), payload...)
	sendData(addr, request)
}


//-------------------------used in client.go----------------
//--------leader------------------//

// HandleCommit rx commit
func HandleCoSiCommit(addr string, request []byte) {
	var buff bytes.Buffer
	var payload cosi.Commitment
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	cosiCommitCh <- commitInfoCh{addr, payload}
}



// HandleCoSiResponse rx response
func HandleCoSiResponse(addr string, request[]byte) {
	var buff bytes.Buffer
	var payload cosi.SignaturePart

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	cosiResponseCh <- responseInfoCh{addr, payload}
}



//--------member------------------//
// HandleAnnounce rx announce
func HandleCoSiAnnounce(request []byte)  {
	var buff bytes.Buffer
	var payload []byte

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	CoSiFlag = true
	cosiAnnounceCh <- payload
}


// HandleCoSiChallenge rx challenge
func HandleCoSiChallenge(request []byte)  {
	var buff bytes.Buffer
	var payload challengeMessage

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	cosiChallengeCh <- payload

}


// HandleCosiSig rx cosisig
func HandleCoSiSig(request []byte) {
	var buff bytes.Buffer
	var payload cosi.SignaturePart

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	cosiSigCh <- payload
}