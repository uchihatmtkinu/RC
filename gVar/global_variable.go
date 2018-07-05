package gVar

//MagicNumber magic
const MagicNumber byte = 66

//ShardSize is the number of miners in one shard
const ShardSize uint32 = 2

//ShardCnt is the number of shards
const ShardCnt uint32 = 2

//used in rep calculation, scaling factor
const RepTP = 1
const RepTN = 1
const RepFP = 1
const RepFN = 1

//channel

const SlidingWindows = 10

//NumTxBlockPerEpoch is the number of txblocks in one epoch
const NumTxBlockPerEpoch = 1

//NumTxListPerEpoch is the number of txblocks in one epoch
const NumTxListPerEpoch = 1

//NumTxBlockForRep is the number of blocks for one rep block
const NumTxBlockForRep = 2
