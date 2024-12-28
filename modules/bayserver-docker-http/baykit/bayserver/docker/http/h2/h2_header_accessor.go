package h2

import "bayserver-core/baykit/bayserver/protocol"

type H2HeaderAccessor struct {
	*protocol.PacketPartAccessor
}

func NewH2HeaderAccessor(pkt protocol.Packet, start int, maxLen int) *H2HeaderAccessor {
	return &H2HeaderAccessor{
		protocol.NewPacketPartAccessor(pkt, start, maxLen),
	}
}

func (a *H2HeaderAccessor) PutInt24(val int) {
	bytes := []byte{
		byte((val >> 16) & 0xFF),
		byte((val >> 8) & 0xFF),
		byte(val & 0xFF),
	}
	a.PutBytes(bytes, 0, 3)
}
