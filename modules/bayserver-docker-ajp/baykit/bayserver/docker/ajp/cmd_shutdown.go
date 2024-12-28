package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 * Shutdown command format
 *
 *   none
 */

type CmdShutdown struct {
	*AjpCommandBase
}

func NewCmdShutdown() *CmdShutdown {
	c := CmdShutdown{
		AjpCommandBase: NewAjpCommandBase(AJP_TYPE_SHUTDOWN, true),
	}
	var _ protocol.Command = &c // cast check
	var _ AjpCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdShutdown) Unpack(pkt protocol.Packet) exception2.IOException {
	c.AjpCommandBase.Unpack(pkt.(*AjpPacket))
	return nil
}

func (c *CmdShutdown) Pack(pkt protocol.Packet) exception2.IOException {
	// must be called from last line
	c.AjpCommandBase.Pack(pkt.(*AjpPacket))
	return nil
}

func (c *CmdShutdown) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(AjpCommandHandler).HandleShutdown(c)
}
