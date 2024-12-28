package fcgi

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
)

type FcgProtocolHandler interface {
	protocol.ProtocolHandler
}

type FcgProtocolHandlerImpl struct {
	impl.ProtocolHandlerImpl
}

func NewFcgProtocolHandler(
	packetUnpacker protocol.PacketUnpacker,
	packetPacker protocol.PacketPacker,
	commandUnpacker protocol.CommandUnpacker,
	commandPacker protocol.CommandPacker,
	commandHandler protocol.CommandHandler,
	serverMode bool) *FcgProtocolHandlerImpl {

	ph := FcgProtocolHandlerImpl{}
	ph.ProtocolHandlerImpl.ConstructProtocolHandler(packetUnpacker, packetPacker, commandUnpacker, commandPacker, commandHandler, serverMode)
	return &ph
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *FcgProtocolHandlerImpl) Reset() {
	h.ProtocolHandlerImpl.Reset()
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (h *FcgProtocolHandlerImpl) Protocol() string {
	return FCG_PROTO_NAME
}

func (h *FcgProtocolHandlerImpl) MaxReqPacketDataSize() int {
	return FCG_PACKET_MAXLEN
}
func (h *FcgProtocolHandlerImpl) MaxResPacketDataSize() int {
	return FCG_PACKET_MAXLEN
}
