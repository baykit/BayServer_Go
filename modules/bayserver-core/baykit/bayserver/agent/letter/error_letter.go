package letter

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/util/exception"
)

type ErrorLetter struct {
	LetterStruct
	Err exception.Exception
}

func NewErrorLetter(st *common.RudderState, err exception.Exception) *ErrorLetter {
	return &ErrorLetter{
		LetterStruct: LetterStruct{
			state: st,
		},
		Err: err,
	}
}
