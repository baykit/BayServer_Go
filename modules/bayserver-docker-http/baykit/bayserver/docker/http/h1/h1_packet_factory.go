package h1

import "bayserver-core/baykit/bayserver/protocol"

func H1PacketFactory(typ int) protocol.Packet {
	return NewH1Packet(typ)
}
