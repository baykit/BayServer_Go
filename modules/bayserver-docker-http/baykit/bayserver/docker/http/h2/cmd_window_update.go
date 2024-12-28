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

type CmdWindowUpdate struct {
	*H2CommandBase
	WindowSizeIncrement int
}

func NewCmdWindowUpdate(streamId int, flags *H2Flags) *CmdWindowUpdate {
	c := &CmdWindowUpdate{
		H2CommandBase: NewH2CommandBase(H2_TYPE_WINDOW_UPDATE, flags, streamId),
	}

	var _ protocol.Command = c // implement check
	var _ H2Command = c        // implement check
	return c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdWindowUpdate) Unpack(pkt protocol.Packet) exception.IOException {
	c.H2CommandBase.Unpack(pkt.(*H2Packet))
	acc := pkt.NewDataAccessor()

	val := acc.GetInt()
	c.WindowSizeIncrement = ExtractInt31(val)
	return nil
}

func (c *CmdWindowUpdate) Pack(pkt protocol.Packet) exception.IOException {
	acc := pkt.NewDataAccessor()
	acc.PutInt(ConsolidateFlagAndInt32(0, c.WindowSizeIncrement))
	c.H2CommandBase.Pack(pkt.(*H2Packet))
	return nil
}

func (c *CmdWindowUpdate) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H2CommandHandler).HandleWindowUpdate(c)
}
