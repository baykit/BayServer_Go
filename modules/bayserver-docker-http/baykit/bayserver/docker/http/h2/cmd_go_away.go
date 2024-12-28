package h2

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

/**
 * HTTP/2 GoAway payload format
 *
 * +-+-------------------------------------------------------------+
 * |R|                  Last-Stream-ID (31)                        |
 * +-+-------------------------------------------------------------+
 * |                      Error Code (32)                          |
 * +---------------------------------------------------------------+
 * |                  Additional Debug Data (*)                    |
 * +---------------------------------------------------------------+
 *
 */

type CmdGoAway struct {
	*H2CommandBase
	LastStreamId int
	ErrorCode    int
	DebugData    []byte
}

func NewCmdGoAway(streamId int, flags *H2Flags) *CmdGoAway {
	c := &CmdGoAway{
		H2CommandBase: NewH2CommandBase(H2_TYPE_GOAWAY, flags, streamId),
	}
	var _ protocol.Command = c // implement check
	var _ H2Command = c        // implement check
	return c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdGoAway) Unpack(pkt protocol.Packet) exception.IOException {
	c.H2CommandBase.Unpack(pkt.(*H2Packet))
	acc := pkt.NewDataAccessor()
	val := acc.GetInt()
	c.LastStreamId = ExtractInt31(val)
	c.ErrorCode = acc.GetInt()
	c.DebugData = make([]byte, pkt.DataLen()-acc.Pos())
	acc.GetBytes(c.DebugData, 0, len(c.DebugData))
	return nil
}

func (c *CmdGoAway) Pack(pkt protocol.Packet) exception.IOException {
	acc := pkt.NewDataAccessor()
	acc.PutInt(c.LastStreamId)
	acc.PutInt(c.ErrorCode)
	if c.DebugData != nil {
		acc.PutBytes(c.DebugData, 0, len(c.DebugData))
	}
	c.H2CommandBase.Pack(pkt.(*H2Packet))
	return nil
}

func (c *CmdGoAway) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H2CommandHandler).HandleGoAway(c)
}
