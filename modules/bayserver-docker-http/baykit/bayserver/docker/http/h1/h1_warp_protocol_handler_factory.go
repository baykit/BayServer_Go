package h1

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
)

var H1WarpProtocolHandlerFactory = func(pktStore *packetstore.PacketStore) protocol.ProtocolHandler {
	warpHandler := NewH1WarpHandler()
	commandUnpacker := NewH1CommandUnpacker(warpHandler, false)
	packetUnpacker := NewH1PacketUnpacker(commandUnpacker, pktStore)
	packetPacker := impl.NewPacketPacker()
	commandPacker := impl.NewCommandPacker(packetPacker, pktStore)
	protocolHandler := NewH1ProtocolHandler(
		packetUnpacker,
		packetPacker,
		commandUnpacker,
		commandPacker,
		warpHandler,
		true)
	warpHandler.Init(protocolHandler)
	return protocolHandler
}
