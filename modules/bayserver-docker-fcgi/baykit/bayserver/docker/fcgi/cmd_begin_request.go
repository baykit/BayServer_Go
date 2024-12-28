package fcgi

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

/**
 * FCGI spec
 *   http://www.mit.edu/~yandros/doc/specs/fcgi-spec.html
 *
 * Begin request command format
 *         typedef struct {
 *             unsigned char roleB1;
 *             unsigned char roleB0;
 *             unsigned char flags;
 *             unsigned char reserved[5];
 *         } FCGI_BeginRequestBody;
 */

const FCGI_KEEP_CON = 1
const FCGI_RESPONDER = 1
const FCGI_AUTHORIZER = 2
const FCGI_FILTER = 3

type CmdBeginRequest struct {
	*FcgCommandBase
	Role    int
	KeepCon bool
}

func NewCmdBeginRequest(reqId int) *CmdBeginRequest {
	c := CmdBeginRequest{
		FcgCommandBase: NewFcgCommandBase(FCG_TYPE_BEGIN_REQUEST, reqId),
		Role:           0,
		KeepCon:        false,
	}
	var _ protocol.Command = &c // cast check
	var _ FcgCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdBeginRequest) Unpack(pkt protocol.Packet) exception2.IOException {
	c.FcgCommandBase.Unpack(pkt.(*FcgPacket))

	acc := pkt.NewDataAccessor()
	c.Role = acc.GetShort()
	flags := acc.GetByte()
	c.KeepCon = (flags & FCGI_KEEP_CON) != 0
	return nil
}

func (c *CmdBeginRequest) Pack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.NewDataAccessor()

	acc.PutShort(c.Role)
	if c.KeepCon {
		acc.PutByte(1)
	} else {
		acc.PutByte(0)
	}
	reserved := make([]byte, 5)
	acc.PutBytes(reserved, 0, len(reserved))

	// must be called from last line
	c.FcgCommandBase.Pack(pkt.(*FcgPacket))
	return nil
}

func (c *CmdBeginRequest) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(FcgCommandHandler).HandleBeginRequest(c)
}

/****************************************/
/* Custome method                       */
/****************************************/
