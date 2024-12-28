package h2

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

/**
 */

type CmdPing struct {
	*H2CommandBase
	opaqueData []byte
}

func NewCmdPing(streamId int, flags *H2Flags, opaqueData []byte) protocol.Command {
	c := &CmdPing{
		H2CommandBase: NewH2CommandBase(H2_TYPE_PING, flags, streamId),
	}
	c.opaqueData = make([]byte, 8)
	if opaqueData != nil {
		copy(c.opaqueData, opaqueData)
	}

	var _ protocol.Command = c // implement check
	var _ H2Command = c        // implement check
	return c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdPing) Unpack(pkt protocol.Packet) exception.IOException {
	c.H2CommandBase.Unpack(pkt.(*H2Packet))
	acc := pkt.NewDataAccessor()
	acc.GetBytes(c.opaqueData, 0, 8)
	return nil
}

func (c *CmdPing) Pack(pkt protocol.Packet) exception.IOException {
	acc := pkt.NewDataAccessor()
	acc.PutBytes(c.opaqueData, 0, len(c.opaqueData))
	c.H2CommandBase.Pack(pkt.(*H2Packet))
	return nil
}

func (c *CmdPing) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H2CommandHandler).HandlePing(c)
}
