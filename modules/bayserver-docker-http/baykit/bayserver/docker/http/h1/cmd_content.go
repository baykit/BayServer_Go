package h1

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/util/exception"
)

type CmdContent struct {
	*impl.CommandBase
	buffer []byte
	start  int
	length int
}

func NewCmdContent(buf []byte, start int, length int) protocol.Command {
	c := CmdContent{
		CommandBase: impl.NewCommandBase(H1_TYPE_CONTENT),
		buffer:      buf,
		start:       start,
		length:      length,
	}
	var _ protocol.Command = &c // cast check
	var _ H1Command = &c        // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdContent) Unpack(pkt protocol.Packet) exception.IOException {
	c.buffer = pkt.Buf()
	c.start = pkt.HeaderLen()
	c.length = pkt.DataLen()
	return nil
}

func (c *CmdContent) Pack(pkt protocol.Packet) exception.IOException {
	acc := pkt.NewDataAccessor()
	acc.PutBytes(c.buffer, c.start, c.length)
	return nil
}

func (c *CmdContent) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception.IOException) {
	return h.(H1CommandHandler).HandleContent(c)
}
