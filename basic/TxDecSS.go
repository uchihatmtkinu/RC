package basic

//Build implements the TxDecSS type given the TxDecSet data
func (a *TxDecSS) Build(b *[]TxDecSet) {
	a.ShardNum = uint32(len(*b))
	a.Header = make([]TDSHeader, a.ShardNum)
	tmp := make(map[[32]byte]uint32, 5000)
	a.TxCnt = 0
	a.Tx = make([][32]byte, 0, 4000)
	for i := uint32(0); i < a.ShardNum; i++ {
		a.Header[i].ID = (*b)[i].ID
		a.Header[i].HashID = (*b)[i].HashID
		a.Header[i].PrevHash = (*b)[i].PrevHash
		a.Header[i].MemCnt = (*b)[i].MemCnt
		a.Header[i].Sig = (*b)[i].Sig
		a.Header[i].TxCnt = (*b)[i].TxCnt
		a.Header[i].MemD = make([]TxDPure, a.Header[i].MemCnt)
		a.Header[i].TxIndex = make([]uint32, a.Header[i].TxCnt)
		for j := uint32(0); j < (*b)[i].MemCnt; j++ {
			a.Header[i].MemD[j].ID = (*b)[i].MemD[j].ID
			a.Header[i].MemD[j].Decision = (*b)[i].MemD[j].Decision
			a.Header[i].MemD[j].Sig = (*b)[i].MemD[j].Sig
		}
		for j := uint32(0); j < (*b)[i].TxCnt; j++ {
			tmpValue, ok := tmp[(*b)[i].TxArray[j]]
			if !ok {
				a.Tx = append(a.Tx, (*b)[i].TxArray[j])
				a.TxCnt++
				tmp[(*b)[i].TxArray[j]] = a.TxCnt
				a.Header[i].TxIndex[j] = a.TxCnt - 1
			} else {
				a.Header[i].TxIndex[j] = tmpValue - 1
			}
		}
	}
}
