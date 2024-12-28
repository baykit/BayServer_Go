package fcgi

import "bayserver-core/baykit/bayserver/protocol"

func FcgPacketFactory(typ int) protocol.Packet {
	return NewFcgPacket(typ)
}
