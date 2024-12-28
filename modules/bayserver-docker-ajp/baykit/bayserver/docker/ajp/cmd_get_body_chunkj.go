package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 * Get Body Chunk format
 *
 * AJP13_GET_BODY_CHUNK :=
 *   prefix_code       6
 *   requested_length  (integer)
 */

type CmdGetBodyChunk struct {
	*AjpCommandBase
	ReqLen int
}

func NewCmdGetBodyChunk() *CmdGetBodyChunk {
	c := CmdGetBodyChunk{
		AjpCommandBase: NewAjpCommandBase(AJP_TYPE_GET_BODY_CHUNK, false),
		ReqLen:         0,
	}
	var _ protocol.Command = &c // cast check
	var _ AjpCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdGetBodyChunk) Unpack(pkt protocol.Packet) exception2.IOException {
	c.AjpCommandBase.Unpack(pkt.(*AjpPacket))
	return nil
}

func (c *CmdGetBodyChunk) Pack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.(*AjpPacket).NewAjpDataAccessor()

	acc.PutByte(c.Type())
	acc.PutShort(c.ReqLen)

	// must be called from last line
	c.AjpCommandBase.Pack(pkt.(*AjpPacket))
	return nil
}

func (c *CmdGetBodyChunk) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(AjpCommandHandler).HandleGetBodyChunk(c)
}
