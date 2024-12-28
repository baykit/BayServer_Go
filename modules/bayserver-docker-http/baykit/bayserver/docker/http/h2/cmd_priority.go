package h2

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

/**
 * HTTP/2 Priority payload format
 *
 * +-+-------------------------------------------------------------+
 * |E|                  Stream Dependency (31)                     |
 * +-+-------------+-----------------------------------------------+
 * |   Weight (8)  |
 * +-+-------------+
 *
 */

type CmdPriority struct {
	*H2CommandBase
	Weight           int
	Excluded         bool
	StreamDependency int
}

func NewCmdPriority(streamId int, flags *H2Flags) protocol.Command {
	c := &CmdPriority{
		H2CommandBase: NewH2CommandBase(H2_TYPE_PRIORITY, flags, streamId),
	}

	var _ protocol.Command = c // implement check
	var _ H2Command = c        // implement check
	return c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdPriority) Unpack(pkt protocol.Packet) exception.IOException {
	c.H2CommandBase.Unpack(pkt.(*H2Packet))
	acc := pkt.NewDataAccessor()

	val := acc.GetInt()
	c.Excluded = ExtractFlag(val) == 1
	c.StreamDependency = ExtractInt31(val)

	c.Weight = acc.GetByte()
	return nil
}

func (c *CmdPriority) Pack(pkt protocol.Packet) exception.IOException {
	acc := pkt.NewDataAccessor()
	acc.PutInt(MakeStreamDependency32(c.Excluded, c.StreamDependency))
	acc.PutByte(c.Weight)
	c.H2CommandBase.Pack(pkt.(*H2Packet))
	return nil
}

func (c *CmdPriority) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H2CommandHandler).HandlePriority(c)
}
