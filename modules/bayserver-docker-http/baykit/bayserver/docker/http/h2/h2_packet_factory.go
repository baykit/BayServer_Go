package h2

import "bayserver-core/baykit/bayserver/protocol"

func H2PacketFactory(typ int) protocol.Packet {
	return NewH2Packet(typ)
}
