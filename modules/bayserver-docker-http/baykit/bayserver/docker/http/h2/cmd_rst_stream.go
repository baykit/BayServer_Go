package h2

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

/**
* HTTP/2 RstStream payload format
*
+---------------------------------------------------------------+
|                        Error Code (32)                        |
+---------------------------------------------------------------+
*
*/

type CmdRstStream struct {
	*H2CommandBase
	ErrorCode int
}

func NewCmdRstStream(streamId int, flags *H2Flags) protocol.Command {
	c := &CmdRstStream{
		H2CommandBase: NewH2CommandBase(H2_TYPE_RST_STREAM, flags, streamId),
	}

	var _ protocol.Command = c // implement check
	var _ H2Command = c        // implement check
	return c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdRstStream) Unpack(pkt protocol.Packet) exception.IOException {
	c.H2CommandBase.Unpack(pkt.(*H2Packet))
	acc := pkt.NewDataAccessor()

	c.ErrorCode = acc.GetInt()
	return nil
}

func (c *CmdRstStream) Pack(pkt protocol.Packet) exception.IOException {
	acc := pkt.NewDataAccessor()
	acc.PutInt(c.ErrorCode)
	c.H2CommandBase.Pack(pkt.(*H2Packet))
	return nil
}

func (c *CmdRstStream) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H2CommandHandler).HandleRstStream(c)
}
