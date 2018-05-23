package basic

//Build implements the TxDecSS type given the TxDecSet data
func (a *TxDecSS) Build(b *[]TxDecSet) {
	a.ShardNum = uint32(len(*b))
	a.Header = make([]TDSHeader, a.ShardNum)
	//var tmp [][32]byte
	for i := uint32(0); i < a.ShardNum; i++ {
		a.Header[i].ID = (*b)[i].ID
		a.Header[i].HashID = (*b)[i].HashID
		a.Header[i].PrevHash = (*b)[i].PrevHash
		a.Header[i].MemCnt = (*b)[i].MemCnt
		a.Header[i].Sig = (*b)[i].Sig
		if i == uint32(0) {
			//tmp = (*b)[i].TxArray
		} else {
			//now := 0
			for j := 0; j < len((*b)[i].TxArray); j++ {

			}
		}
	}

}
