package letter

import (
	"bayserver-core/baykit/bayserver/common"
)

type WroteLetter struct {
	LetterStruct
	NBytes int
}

func NewWroteLetter(st *common.RudderState, n int) *WroteLetter {
	return &WroteLetter{
		LetterStruct: LetterStruct{
			state: st,
		},
		NBytes: n,
	}
}
