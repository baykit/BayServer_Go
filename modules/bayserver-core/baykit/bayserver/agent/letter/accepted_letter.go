package letter

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/rudder"
)

type AcceptedLetter struct {
	LetterStruct
	ClientRudder rudder.Rudder
}

func NewAcceptedLetter(st *common.RudderState, clientRd rudder.Rudder) Letter {
	return &AcceptedLetter{
		LetterStruct: LetterStruct{
			state: st,
		},
		ClientRudder: clientRd,
	}
}
