package Reputation

import (
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
	"sync"
)
//RepPowRxInfo receive pow info
type RepPowInfo struct {
	Round	int
	Nonce	int
	Hash 	[32]byte
}
//channel used in rep pow
//RepPowRxCh rx pow repblock from others
var RepPowRxCh chan RepPowInfo
//RepPowTxCh tx a pow repblock
var RepPowTxCh chan RepPowInfo
//RepPowRxValidate flag - validate the received repblock
var RepPowRxValidate chan bool
//MyRepBlockChain my reputation blockchain
var MyRepBlockChain *RepBlockchain
//RepBlockIter an iterator on repblockchain
var RepBlockChainIter	*RepBlockchainIterator
//SafeSyncBlock used in sync block
type SafeSyncBlock struct {
	Block		*SyncBlock
	Epoch		int
	Mu			sync.RWMutex
}
//SafeSyncBlock used in sync block
type SafeRepBlock struct {
	Block		*RepBlock
	Round 		int
	Mu			sync.RWMutex
}
//CurrentSyncBlock current sync block
var CurrentSyncBlock	SafeSyncBlock
//CurrentSyncBlock current sync block
var CurrentRepBlock		SafeRepBlock
//CurrentCoSignature current cosinature
var CurrentCoSignature	cosi.SignaturePart