package h2

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

/**
 * HTTP/2 Data payload format
 *
 * +---------------+
 * |Pad Length? (8)|
 * +---------------+-----------------------------------------------+
 * |                            Data (*)                         ...
 * +---------------------------------------------------------------+
 * |                           Padding (*)                       ...
 * +---------------------------------------------------------------+
 */

type CmdData struct {
	*H2CommandBase
	data   []byte
	start  int
	length int
}

func NewCmdData(streamId int, flags *H2Flags, data []byte, start int, length int) *CmdData {
	c := &CmdData{
		H2CommandBase: NewH2CommandBase(H2_TYPE_DATA, flags, streamId),
		data:          data,
		start:         start,
		length:        length,
	}
	var _ protocol.Command = c // implement check
	var _ H2Command = c        // implement check
	return c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdData) Unpack(pkt protocol.Packet) exception.IOException {
	c.H2CommandBase.Unpack(pkt.(*H2Packet))
	c.data = pkt.Buf()
	c.start = pkt.HeaderLen()
	c.length = pkt.DataLen()
	return nil
}

func (c *CmdData) Pack(pkt protocol.Packet) exception.IOException {
	acc := pkt.NewDataAccessor()
	if c.Flags().IsPadded() {
		bayserver.FatalError(exception.NewSink("Padding not supported"))
	}
	acc.PutBytes(c.data, c.start, c.length)
	c.H2CommandBase.Pack(pkt.(*H2Packet))
	return nil
}

func (c *CmdData) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H2CommandHandler).HandleData(c)
}
