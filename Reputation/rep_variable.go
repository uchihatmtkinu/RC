package Reputation

import "github.com/uchihatmtkinu/RC/Reputation/cosi"

//channel used in rep pow
//RepPowRxCh rx pow repblock from others
var RepPowRxCh chan *RepBlock
//RepPowTxCh tx a pow repblock
var RepPowTxCh chan RepBlock
//RepPowRxValidate flag - validate the received repblock
var RepPowRxValidate chan bool
//MyRepBlockChain my reputation blockchain
var MyRepBlockChain *RepBlockchain
//RepBlockIter an iterator on repblockchain
var RepBlockChainIter	*RepBlockchainIterator
//CurrentSyncBlock current sync block
var CurrentSyncBlock	*SyncBlock
//CurrentCoSignature current cosinature
var CurrentCoSignature	cosi.SignaturePart