package Reputation




//channel used in rep pow
var RepPowRxCh chan RepBlock
var RepPowTxCh chan RepBlock
var RepPowRxValidate chan bool

var MyRepBlockChain RepBlockchain