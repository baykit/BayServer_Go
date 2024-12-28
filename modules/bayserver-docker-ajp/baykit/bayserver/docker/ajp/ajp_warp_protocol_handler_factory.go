package ajp

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
)

var AjpWarpProtocolHandlerFactory = func(pktStore *packetstore.PacketStore) protocol.ProtocolHandler {
	warpHandler := NewAjpWarpHandler()
	commandUnpacker := NewAjpCommandUnpacker(warpHandler)
	packetUnpacker := NewAjpPacketUnpacker(commandUnpacker, pktStore)
	packetPacker := impl.NewPacketPacker()
	commandPacker := impl.NewCommandPacker(packetPacker, pktStore)
	protocolHandler := NewAjpProtocolHandler(
		packetUnpacker,
		packetPacker,
		commandUnpacker,
		commandPacker,
		warpHandler,
		true)
	warpHandler.Init(protocolHandler)
	return protocolHandler
}
