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

type CmdStdIn struct {
	*InOutCommandBase
}

func NewCmdStdIn(reqId int) *CmdStdIn {
	c := CmdStdIn{
		NewInOutCommandBase(FCG_TYPE_STDIN, reqId, nil, 0, 0),
	}
	var _ protocol.Command = &c // cast check
	var _ FcgCommand = &c       // cast check
	return &c
}

func NewCmdStdIn2(reqId int, data []byte, start int, length int) *CmdStdIn {
	c := CmdStdIn{
		NewInOutCommandBase(FCG_TYPE_STDIN, reqId, data, start, length),
	}
	var _ protocol.Command = &c // cast check
	var _ FcgCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdStdIn) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(FcgCommandHandler).HandleStdIn(c)
}
