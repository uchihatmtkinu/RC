package Reputation




//channel used in rep pow
//RepPowRxCh rx pow repblock from others
var RepPowRxCh chan RepBlock
//RepPowTxCh tx a pow repblock
var RepPowTxCh chan RepBlock
//RepPowRxValidate flag - validate the received repblock
var RepPowRxValidate chan bool
//MyRepBlockChain my reputation blockchain
var MyRepBlockChain RepBlockchain