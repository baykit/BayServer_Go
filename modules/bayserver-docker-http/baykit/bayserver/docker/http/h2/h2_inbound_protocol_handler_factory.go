package h2

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
)

var H2InboundProtocolHandlerFactory = func(pktStore *packetstore.PacketStore) protocol.ProtocolHandler {
	inboundHandler := NewH2InboundHandler()
	commandUnpacker := NewH2CommandUnpacker(inboundHandler)
	packetUnpacker := NewH2PacketUnpacker(commandUnpacker, pktStore, true)
	packetPacker := impl.NewPacketPacker()
	commandPacker := impl.NewCommandPacker(packetPacker, pktStore)
	protocolHandler := NewH2ProtocolHandler(
		packetUnpacker,
		packetPacker,
		commandUnpacker,
		commandPacker,
		inboundHandler,
		true)
	inboundHandler.Init(protocolHandler)
	return protocolHandler
}
