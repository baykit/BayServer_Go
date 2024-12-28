package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 * End response body format
 *
 * AJP13_END_RESPONSE :=
 *   prefix_code       5
 *   reuse             (boolean)
 */

type CmdEndResponse struct {
	*AjpCommandBase
	Reuse bool
}

func NewCmdEndResponse() *CmdEndResponse {
	c := CmdEndResponse{
		AjpCommandBase: NewAjpCommandBase(AJP_TYPE_END_RESPONSE, false),
		Reuse:          false,
	}
	var _ protocol.Command = &c // cast check
	var _ AjpCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdEndResponse) Unpack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.(*AjpPacket).NewAjpDataAccessor()
	acc.GetByte() // prefix code
	c.Reuse = acc.GetByte() != 0
	return nil
}

func (c *CmdEndResponse) Pack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.(*AjpPacket).NewAjpDataAccessor()

	acc.PutByte(c.Type())
	if c.Reuse {
		acc.PutByte(1)
	} else {
		acc.PutByte(0)
	}

	// must be called from last line
	c.AjpCommandBase.Pack(pkt.(*AjpPacket))
	return nil
}

func (c *CmdEndResponse) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(AjpCommandHandler).HandleEndResponse(c)
}
