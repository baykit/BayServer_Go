package letter

import (
	"bayserver-core/baykit/bayserver/common"
)

type ConnectedLetter struct {
	LetterStruct
}

func NewConnectedLetter(st *common.RudderState) *ConnectedLetter {
	return &ConnectedLetter{
		LetterStruct: LetterStruct{
			state: st,
		},
	}
}
