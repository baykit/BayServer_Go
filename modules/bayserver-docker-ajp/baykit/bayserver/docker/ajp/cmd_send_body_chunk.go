package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 * Send body chunk format
 *
 * AJP13_SEND_BODY_CHUNK :=
 *   prefix_code   3
 *   chunk_length  (integer)
 *   chunk        *(byte)
 */

const AJP_MAX_CHUNKLEN = AJP_MAX_DATA_LEN - 4

type CmdSendBodyChunk struct {
	*AjpCommandBase
	Chunk  []byte
	Offset int
	Length int
}

func NewCmdSendBodyChunk(buf []byte, ofs int, length int) *CmdSendBodyChunk {
	c := CmdSendBodyChunk{
		AjpCommandBase: NewAjpCommandBase(AJP_TYPE_SEND_BODY_CHUNK, false),
		Chunk:          buf,
		Offset:         ofs,
		Length:         length,
	}
	var _ protocol.Command = &c // cast check
	var _ AjpCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdSendBodyChunk) Unpack(pkt protocol.Packet) exception2.IOException {
	c.AjpCommandBase.Unpack(pkt.(*AjpPacket))
	acc := pkt.(*AjpPacket).NewAjpDataAccessor()
	acc.GetByte() // code
	c.Length = acc.GetShort()
	if c.Chunk == nil || c.Length > len(c.Chunk) {
		c.Chunk = make([]byte, c.Length)
	}
	acc.GetBytes(c.Chunk, 0, c.Length)
	return nil
}

func (c *CmdSendBodyChunk) Pack(pkt protocol.Packet) exception2.IOException {
	if c.Length > AJP_MAX_CHUNKLEN {
		return exception.NewProtocolException("Illegal argument")
	}
	acc := pkt.(*AjpPacket).NewAjpDataAccessor()

	acc.PutByte(c.Type())
	acc.PutShort(c.Length)
	acc.PutBytes(c.Chunk, c.Offset, c.Length)
	acc.PutByte(0) // maybe document bug

	// must be called from last line
	c.AjpCommandBase.Pack(pkt.(*AjpPacket))
	return nil
}

func (c *CmdSendBodyChunk) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(AjpCommandHandler).HandleSendBodyChunk(c)
}
