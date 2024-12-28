package ajp

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
)

var AjpInboundProtocolHandlerFactory = func(pktStore *packetstore.PacketStore) protocol.ProtocolHandler {
	inboundHandler := NewAjpInboundHandler()
	commandUnpacker := NewAjpCommandUnpacker(inboundHandler)
	packetUnpacker := NewAjpPacketUnpacker(commandUnpacker, pktStore)
	packetPacker := impl.NewPacketPacker()
	commandPacker := impl.NewCommandPacker(packetPacker, pktStore)
	protocolHandler := NewAjpProtocolHandler(
		packetUnpacker,
		packetPacker,
		commandUnpacker,
		commandPacker,
		inboundHandler,
		true)
	inboundHandler.Init(protocolHandler)
	return protocolHandler
}
