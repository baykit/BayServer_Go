package impl

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util/exception"
)

type PacketPackerImpl struct {
}

func NewPacketPacker() *PacketPackerImpl {
	return &PacketPackerImpl{}
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (cp *PacketPackerImpl) Reset() {

}

/****************************************/
/* Implements PacketPacker              */
/****************************************/

func (cp *PacketPackerImpl) Post(sip ship.Ship, pkt protocol.Packet, lis common.DataConsumeListener) exception.IOException {
	return sip.Transporter().ReqWrite(sip.Rudder(), pkt.Buf()[0:pkt.BufLen()], nil, pkt, func() { lis() })
}
