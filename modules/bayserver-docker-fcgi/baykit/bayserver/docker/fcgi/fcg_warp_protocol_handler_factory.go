package fcgi

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
)

var FcgWarpProtocolHandlerFactory = func(pktStore *packetstore.PacketStore) protocol.ProtocolHandler {
	warpHandler := NewFcgWarpHandler()
	commandUnpacker := NewFcgCommandUnpacker(warpHandler)
	packetUnpacker := NewFcgPacketUnpacker(commandUnpacker, pktStore)
	packetPacker := impl.NewPacketPacker()
	commandPacker := impl.NewCommandPacker(packetPacker, pktStore)
	protocolHandler := NewFcgProtocolHandler(
		packetUnpacker,
		packetPacker,
		commandUnpacker,
		commandPacker,
		warpHandler,
		true)
	warpHandler.Init(protocolHandler)
	return protocolHandler
}
