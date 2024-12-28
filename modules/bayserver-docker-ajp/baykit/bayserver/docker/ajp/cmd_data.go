package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 * Data command format
 *
 *   raw data
 */

type CmdData struct {
	*AjpCommandBase
	Start  int
	Length int
	Data   []byte
}

func NewCmdData(data []byte, start int, length int) *CmdData {
	c := CmdData{
		AjpCommandBase: NewAjpCommandBase(AJP_TYPE_DATA, true),
		Start:          start,
		Length:         length,
		Data:           data,
	}
	var _ protocol.Command = &c // cast check
	var _ AjpCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdData) Unpack(pkt protocol.Packet) exception2.IOException {
	c.AjpCommandBase.Unpack(pkt.(*AjpPacket))

	acc := pkt.(*AjpPacket).NewAjpDataAccessor()
	c.Length = acc.GetShort()
	c.Data = pkt.Buf()
	c.Start = pkt.HeaderLen() + 2
	return nil
}

func (c *CmdData) Pack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.(*AjpPacket).NewAjpDataAccessor()

	acc.PutShort(c.Length)
	acc.PutBytes(c.Data, c.Start, c.Length)

	// must be called from last line
	c.AjpCommandBase.Pack(pkt.(*AjpPacket))
	return nil
}

func (c *CmdData) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(AjpCommandHandler).HandleData(c)
}
