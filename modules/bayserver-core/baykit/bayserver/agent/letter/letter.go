package letter

import (
	"bayserver-core/baykit/bayserver/common"
)

type LetterType int8

type Letter interface {
	State() *common.RudderState
}

type LetterStruct struct {
	state *common.RudderState
}

func (l *LetterStruct) State() *common.RudderState {
	return l.state
}
