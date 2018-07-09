package gVar

//MagicNumber magic
const MagicNumber byte = 66

//ShardSize is the number of miners in one shard
const ShardSize uint32 = 100

//ShardCnt is the number of shards
const ShardCnt uint32 = 1

//used in rep calculation, scaling factor
const RepTP = 1
const RepTN = 1
const RepFP = 1
const RepFN = 1

//channel

const SlidingWindows = 10

//NumTxBlockPerEpoch is the number of txblocks in one epoch
const NumTxPerEpoch = 24000 //48000

//NumTxListPerEpoch is the number of txblocks in one epoch
const NumTxListPerEpoch = 12 //60

//NumTxBlockForRep is the number of blocks for one rep block
const NumTxBlockForRep = 3 //10

const NumTxPerBlock = 2000 //2000

const NumTxPerTL = 2000 //400

//const GensisAcc = []byte{0}

const GensisAccValue = 2147483647

const TxSendInterval = 1

const NumOfTxForTest = 24000 //int(2 * 60 * 4000 * ShardSize)

const GeneralSleepTime = 100
