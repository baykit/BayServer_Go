package letter

import (
	"bayserver-core/baykit/bayserver/common"
)

type ReadLetter struct {
	LetterStruct
	NBytes  int
	Address string
}

func NewReadLetter(st *common.RudderState, n int, adr string) *ReadLetter {
	return &ReadLetter{
		LetterStruct: LetterStruct{
			state: st,
		},
		NBytes:  n,
		Address: adr,
	}
}
