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
 * Endrequest command format
 *         typedef struct {
 *             unsigned char appStatusB3;
 *             unsigned char appStatusB2;
 *             unsigned char appStatusB1;
 *             unsigned char appStatusB0;
 *             unsigned char protocolStatus;
 *             unsigned char reserved[3];
 *         } FCGI_EndRequestBody;
 */

const FCGI_REQUEST_COMPLETE = 0
const FCGI_CANT_MPX_CONN = 1
const FCGI_OVERLOADED = 2
const FCGI_UNKNOWN_ROLE = 3

type CmdEndRequest struct {
	*FcgCommandBase
	AppStatus      int
	ProtocolStatus int
}

func NewCmdEndRequest(reqId int) *CmdEndRequest {
	c := CmdEndRequest{
		FcgCommandBase: NewFcgCommandBase(FCG_TYPE_END_REQUEST, reqId),
		AppStatus:      0,
		ProtocolStatus: FCGI_REQUEST_COMPLETE,
	}
	var _ protocol.Command = &c // cast check
	var _ FcgCommand = &c       // cast check
	return &c
}

/****************************************/
/* Implements Command                   */
/****************************************/

func (c *CmdEndRequest) Unpack(pkt protocol.Packet) exception2.IOException {
	c.FcgCommandBase.Unpack(pkt.(*FcgPacket))

	acc := pkt.NewDataAccessor()
	c.AppStatus = acc.GetInt()
	c.ProtocolStatus = acc.GetByte()
	return nil
}

func (c *CmdEndRequest) Pack(pkt protocol.Packet) exception2.IOException {
	acc := pkt.NewDataAccessor()

	acc.PutInt(c.AppStatus)
	acc.PutByte(c.ProtocolStatus)
	reserved := make([]byte, 3)
	acc.PutBytes(reserved, 0, len(reserved))
	// must be called from last line
	c.FcgCommandBase.Pack(pkt.(*FcgPacket))
	return nil
}

func (c *CmdEndRequest) Handle(h protocol.CommandHandler) (common.NextSocketAction, exception2.IOException) {
	return h.(FcgCommandHandler).HandleEndRequest(c)
}

/****************************************/
/* Custome method                       */
/****************************************/
