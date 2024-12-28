package letter

import (
	"bayserver-core/baykit/bayserver/common"
)

type ClosedLetter struct {
	LetterStruct
}

func NewClosedLetter(st *common.RudderState) *ClosedLetter {
	return &ClosedLetter{
		LetterStruct: LetterStruct{
			state: st,
		},
	}
}
