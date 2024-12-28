package fcgi

import (
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 * FCGI spec
 *   http://www.mit.edu/~yandros/doc/specs/fcgi-spec.html
 *
 * StdIn/StdOut/StdErr command format
 *   raw data
 */

const MAX_DATA_LEN = FCG_PACKET_MAXLEN - FCG_PREAMBLE_SIZE

type InOutCommandBase struct {
	*FcgCommandBase
	Start  int
	Length int
	Data   []byte
}

func NewInOutCommandBase(typ int, reqId int, data []byte, start int, length int) *InOutCommandBase {
	c := InOutCommandBase{
		FcgCommandBase: NewFcgCommandBase(typ, reqId),
		Data:           data,
		Start:          start,
		Length:         length,
	}
	var _ FcgCommand = c // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *InOutCommandBase) Unpack(pkt protocol.Packet) exception2.IOException {
	c.FcgCommandBase.Unpack(pkt.(*FcgPacket))

	c.Start = pkt.HeaderLen()
	c.Length = pkt.DataLen()
	c.Data = pkt.Buf()
	return nil
}

func (c *InOutCommandBase) Pack(pkt protocol.Packet) exception2.IOException {

	if c.Data != nil && len(c.Data) > 0 {
		acc := pkt.NewDataAccessor()
		acc.PutBytes(c.Data, c.Start, c.Length)
	}

	// must be called from last line
	c.FcgCommandBase.Pack(pkt.(*FcgPacket))
	return nil
}

/****************************************/
/* Custome method                       */
/****************************************/
