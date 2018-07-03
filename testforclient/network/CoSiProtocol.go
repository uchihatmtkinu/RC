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
	// cosi begin
	fmt.Println("Leader CoSi")
	CoSiFlag = true
	//To simplify the problem, we just validate the previous repblock hash
	sbMessage = prevRepBlockHash[:]
	commits = make([]cosi.Commitment, int(gVar.ShardSize))
	pubKeys = make([]ed25519.PublicKey, int(gVar.ShardSize))
	//priKeys := make([]ed25519.PrivateKey, int(gVar.ShardSize))
	s1 := [64]byte{1}
	myCommit, mySecret, _ = cosi.Commit(bytes.NewReader(s1[:]))

	//byte mask 0-7 bit in one byte represent user 0-7, 8-15...
	//cosimask used in cosi announce, indicate the number of users sign the block.
	//responsemask, used in cosi, leader resent the order to the member have signed the block
	intilizeMaskBit(&cosimask, (int(gVar.ShardSize)+7)>>3,cosi.Disabled)
	intilizeMaskBit(&responsemask, (int(gVar.ShardSize)+7)>>3,cosi.Disabled)

	//handle leader's commit
	cosiCommitCh = make(chan commitInfo, bufferSize)
	commits[shard.MyMenShard.InShardId] = myCommit
	setMaskBit(shard.MyMenShard.InShardId, cosi.Enabled, &cosimask)
	setMaskBit(shard.MyMenShard.InShardId, cosi.Enabled, &responsemask)

	//sent announcement
	for i:=0; i < int(gVar.ShardSize); i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		pubKeys[it.InShardId] = it.CosiPub
		//priKeys[it.InShardId] = it.RealAccount.CosiPri
		if i!=0 {
			SendCosiMessage(it.Address, "cosiAnnoun", sbMessage)
		}
	}
	fmt.Println("sent CoSi announce")

	//handle members' commits
	signCount := 1
	timeoutflag := true
	for timeoutflag && signCount < int (gVar.ShardSize){
		select {
		case commitMessage := <-cosiCommitCh:
			commits[(*ms)[commitMessage.ID].InShardId] = commitMessage.Commit
			setMaskBit((*ms)[commitMessage.ID].InShardId, cosi.Enabled, &cosimask)
			signCount++
			fmt.Println("Received commit from Global ID: ",commitMessage.ID,", commits count:",signCount,"/",int(gVar.ShardSize))
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

	currentChaMessage := challengeInfo{aggregatePublicKey, aggregateCommit}

	//sign or challenge
	cosiResponseCh = make(chan responseInfo, bufferSize)
	for i:=uint32(1); i <gVar.ShardSize; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		if maskBit(it.InShardId, &cosimask)==cosi.Enabled {
			SendCosiMessage(it.Address, "cosiChallen", currentChaMessage)
		}
	}
	fmt.Println("Sent CoSi Challenage")
	//handle response
	sigParts = make([]cosi.SignaturePart, shard.NumMems)

	responseCount := 1
	//timeoutflag = true
	for responseCount < signCount{
		select {
		case reponseMessage := <-cosiResponseCh:
			it = &(*ms)[reponseMessage.ID]
			sigParts[it.InShardId] = reponseMessage.Sig
			setMaskBit(it.InShardId, cosi.Disabled, &responsemask)
			responseCount++
			fmt.Println("Received response from Global ID: ",reponseMessage.ID,", reponses count:",responseCount,"/",signCount)
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
	mySigPart := cosi.Cosign(shard.MyMenShard.RealAccount.CosiPri, mySecret, sbMessage, aggregatePublicKey, aggregateCommit)
	sigParts[shard.MyMenShard.InShardId] = mySigPart

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
	fmt.Println("Add a new sync block.")
	//close CoSi
	CoSiFlag =false
	close(cosiCommitCh)
	close(cosiResponseCh)

	//valid := cosi.Verify(pubKeys, nil, sbMessage, cosiSig)

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
	//cosiAnnounceCh = make(chan []byte)
	cosiChallengeCh = make(chan challengeInfo)
	cosiSigCh = make(chan cosi.SignaturePart)
	CoSiFlag = true
	fmt.Println("Member CoSi")
	//generate pubKeys
	pubKeys = make([]ed25519.PublicKey, shard.NumMems)
	for i:=0; i < shard.NumMems; i++ {
		it = &(*ms)[shard.ShardToGlobal[shard.MyMenShard.Shard][i]]
		pubKeys[it.InShardId] = it.CosiPub
	}

	//receive announce and verify message

	sbMessage = prevRepBlockHash[:]
	leaderSBMessage := <-cosiAnnounceCh
	//close(cosiAnnounceCh)
	if !verifySBMessage(sbMessage, leaderSBMessage) {
		fmt.Println("Sync Block from leader is wrong!")
		//TODO send warning
	}
	fmt.Println("received cosi announce")

	//send commit
	s1 := [64]byte{2}
	myCommit, mySecret, _ = cosi.Commit(bytes.NewReader(s1[:]))
	SendCosiMessage(LeaderAddr, "cosiCommit", commitInfo{MyGlobalID,myCommit})
	fmt.Println("sent cosi commit")
	//receive challenge
	currentChaMessage :=  <- cosiChallengeCh

	fmt.Println("received cosi challenge from leader")
	//send signature

	sigPart := cosi.Cosign(shard.MyMenShard.RealAccount.CosiPri, mySecret, sbMessage, currentChaMessage.AggregatePublicKey, currentChaMessage.AggregateCommit)
	SendCosiMessage(LeaderAddr, "cosiRespon", responseInfo{MyGlobalID, sigPart})

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
func HandleCoSiCommit(request []byte) {
	var buff bytes.Buffer
	var payload commitInfo
	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	cosiCommitCh <- payload
}



// HandleCoSiResponse rx response
func HandleCoSiResponse(request[]byte) {
	var buff bytes.Buffer
	var payload responseInfo

	buff.Write(request)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	cosiResponseCh <- payload
}



//------------------------member------------------//
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

	//TODO modify the sign message
	//go MemberCosiProcess(&shard.GlobalGroupMems,Reputation)
	cosiAnnounceCh <- payload
}


// HandleCoSiChallenge rx challenge
func HandleCoSiChallenge(request []byte)  {
	var buff bytes.Buffer
	var payload challengeInfo

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