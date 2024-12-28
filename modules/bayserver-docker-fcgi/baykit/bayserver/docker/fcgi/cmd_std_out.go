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

type CmdStdOut struct {
	*InOutCommandBase
}

func NewCmdStdOut(reqId int, data []byte, start int, length int) *CmdStdOut {
	c := CmdStdOut{
		NewInOutCommandBase(FCG_TYPE_STDOUT, reqId, data, start, length),
	}
	var _ protocol.Command = &c // cast check
	var _ FcgCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdStdOut) Unpack(pkt protocol.Packet) exception2.IOException {
	return c.InOutCommandBase.Unpack(pkt)
}

func (c *CmdStdOut) Pack(pkt protocol.Packet) exception2.IOException {
	return c.InOutCommandBase.Pack(pkt)
}

func (c *CmdStdOut) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(FcgCommandHandler).HandleStdOut(c)
}

/****************************************/
/* Implements FcgCommand                */
/****************************************/
