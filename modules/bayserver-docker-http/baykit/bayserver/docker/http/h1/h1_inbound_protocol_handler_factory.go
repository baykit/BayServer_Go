package h1

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
)

var H1InboundProtocolHandlerFactory = func(pktStore *packetstore.PacketStore) protocol.ProtocolHandler {
	inboundHandler := NewH1InboundHandler()
	commandUnpacker := NewH1CommandUnpacker(inboundHandler, true)
	packetUnpacker := NewH1PacketUnpacker(commandUnpacker, pktStore)
	packetPacker := impl.NewPacketPacker()
	commandPacker := impl.NewCommandPacker(packetPacker, pktStore)
	protocolHandler := NewH1ProtocolHandler(
		packetUnpacker,
		packetPacker,
		commandUnpacker,
		commandPacker,
		inboundHandler,
		true)
	inboundHandler.Init(protocolHandler)
	return protocolHandler
}
