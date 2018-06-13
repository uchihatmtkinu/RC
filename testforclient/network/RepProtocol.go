package network

import "github.com/uchihatmtkinu/RC/Reputation"

func repProcess(b *Reputation.RepBlock) {
	MyPoW = *Reputation.NewProofOfWork(b)
	MyPoW.Run()

}
