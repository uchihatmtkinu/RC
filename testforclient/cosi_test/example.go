package main

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/ed25519"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
)

// This example demonstrates how to generate a
// collective signature involving two cosigners,
// and how to check the resulting collective signature.
func main() {

	// Create keypairs for the two cosigners.
	pubKey1, priKey1, _ := ed25519.GenerateKey(nil)
	//pubKey2, priKey2, _ := ed25519.GenerateKey(nil)
	pubKeys := []ed25519.PublicKey{pubKey1}

	// Sign a test message.
	message := []byte("Hello World")
	sig := Sign(message, pubKeys, priKey1)

	// Now verify the resulting collective signature.
	// This can be done by anyone any time, not just the leader.
	valid := cosi.Verify(pubKeys, nil, message, sig)
	fmt.Printf("signature valid: %v", valid)

	// Output:
	// signature valid: true
}

// Helper function to implement a bare-bones cosigning process.
// In practice the two cosigners would be on different machines
// ideally managed by independent administrators or key-holders.
func Sign(message []byte, pubKeys []ed25519.PublicKey,
	priKey1 ed25519.PrivateKey) []byte {

	// Each cosigner first needs to produce a per-message commit.
	commit1, secret1, _ := cosi.Commit(nil)
	//commit2, secret2, _ := cosi.Commit(nil)
	commits := []cosi.Commitment{commit1,cosi.Commitment{}}
	fmt.Println(commits)
	// The leader then combines these into an aggregate commit.
	cosigners := cosi.NewCosigners(pubKeys, nil)
	aggregatePublicKey := cosigners.AggregatePublicKey()
	aggregateCommit := cosigners.AggregateCommit(commits)

	// The cosigners now produce their parts of the collective signature.
	sigPart1 := cosi.Cosign(priKey1, secret1, message, aggregatePublicKey, aggregateCommit)
	//sigPart2 := cosi.Cosign(priKey2, secret2, message, aggregatePublicKey, aggregateCommit)
	sigParts := []cosi.SignaturePart{sigPart1}

	// Finally, the leader combines the two signature parts
	// into a final collective signature.
	sig := cosigners.AggregateSignature(aggregateCommit, sigParts)

	return sig
}