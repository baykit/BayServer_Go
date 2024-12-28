package protocol

type PacketFactory func(typ int) Packet
