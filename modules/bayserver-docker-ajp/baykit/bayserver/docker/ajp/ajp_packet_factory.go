package ajp

import "bayserver-core/baykit/bayserver/protocol"

func AjpPacketFactory(typ int) protocol.Packet {
	return NewAjpPacket(typ)
}
