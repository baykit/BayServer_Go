package h2

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
)

/**
 * HTTP/2 Header payload format
 *
 * +---------------+
 * |Pad Length? (8)|
 * +-+-------------+-----------------------------------------------+
 * |E|                 Stream Dependency? (31)                     |
 * +-+-------------+-----------------------------------------------+
 * |  Weight? (8)  |
 * +-+-------------+-----------------------------------------------+
 * |                   Header Block Fragment (*)                 ...
 * +---------------------------------------------------------------+
 * |                           Padding (*)                       ...
 * +---------------------------------------------------------------+
 */

type CmdHeaders struct {
	*H2CommandBase
	PadLength        int
	Excluded         bool
	StreamDependency int
	Weight           int
	HeaderBlocks     []*HeaderBlock
}

func NewCmdHeaders(streamId int, flags *H2Flags) *CmdHeaders {
	c := &CmdHeaders{
		H2CommandBase: NewH2CommandBase(H2_TYPE_HEADERS, flags, streamId),
	}
	var _ protocol.Command = c // implement check
	var _ H2Command = c        // implement check
	return c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdHeaders) Unpack(pkt protocol.Packet) exception.IOException {
	h2pkt := pkt.(*H2Packet)
	c.H2CommandBase.Unpack(h2pkt)

	acc := h2pkt.NewH2DataAccessor()

	if h2pkt.Flags.IsPadded() {
		c.PadLength = acc.GetByte()
	}
	if h2pkt.Flags.IsPriority() {
		val := acc.GetInt()
		c.Excluded = ExtractFlag(val) == 1
		c.StreamDependency = ExtractInt31(val)
		c.Weight = acc.GetByte()
	}
	c.readHeaderBlock(acc, pkt.DataLen())
	return nil
}

func (c *CmdHeaders) Pack(pkt protocol.Packet) exception.IOException {
	h2pkt := pkt.(*H2Packet)

	acc := h2pkt.NewH2DataAccessor()
	if c.Flags().IsPadded() {
		acc.PutByte(c.PadLength)
	}
	if c.Flags().IsPriority() {
		acc.PutInt(MakeStreamDependency32(c.Excluded, c.StreamDependency))
		acc.PutByte(c.Weight)
	}
	c.writeHeaderBlock(acc)
	c.H2CommandBase.Pack(pkt.(*H2Packet))
	return nil
}

func (c *CmdHeaders) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H2CommandHandler).HandleHeaders(c)
}

func (c *CmdHeaders) readHeaderBlock(acc *H2DataAccessor, length int) {
	for acc.Pos() < length {
		blk := HeaderBlockUnPack(acc)
		if baylog.IsTraceMode() {
			baylog.Trace("h2: header block read: %s", blk)
		}
		c.AddHeaderBlock(blk)
	}
}

func (c *CmdHeaders) writeHeaderBlock(acc *H2DataAccessor) {
	for _, blk := range c.HeaderBlocks {
		HeaderBlockPack(blk, acc)
	}
}

func (c *CmdHeaders) AddHeaderBlock(blk *HeaderBlock) {
	c.HeaderBlocks = append(c.HeaderBlocks, blk)
}
