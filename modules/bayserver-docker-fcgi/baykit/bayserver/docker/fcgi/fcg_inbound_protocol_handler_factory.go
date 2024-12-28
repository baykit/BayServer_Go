package fcgi

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
)

var FcgInboundProtocolHandlerFactory = func(pktStore *packetstore.PacketStore) protocol.ProtocolHandler {
	inboundHandler := NewFcgInboundHandler()
	commandUnpacker := NewFcgCommandUnpacker(inboundHandler)
	packetUnpacker := NewFcgPacketUnpacker(commandUnpacker, pktStore)
	packetPacker := impl.NewPacketPacker()
	commandPacker := impl.NewCommandPacker(packetPacker, pktStore)
	protocolHandler := NewFcgProtocolHandler(
		packetUnpacker,
		packetPacker,
		commandUnpacker,
		commandPacker,
		inboundHandler,
		true)
	inboundHandler.Init(protocolHandler)
	return protocolHandler
}
