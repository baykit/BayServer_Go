package fcgi

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 * FCGI spec
 *   http://www.mit.edu/~yandros/doc/specs/fcgi-spec.html
 *
 * StdErr command format
 *   raw data
 */

type CmdStdErr struct {
	*InOutCommandBase
}

func NewCmdStdErr(reqId int) *CmdStdErr {
	c := CmdStdErr{
		NewInOutCommandBase(FCG_TYPE_STDERR, reqId, nil, 0, 0),
	}
	var _ protocol.Command = &c // cast check
	var _ FcgCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdStdErr) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(FcgCommandHandler).HandleStdErr(c)
}

/****************************************/
/* Implements FcgCommand                */
/****************************************/
