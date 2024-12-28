package ajp

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
)

type AjpProtocolHandler interface {
	protocol.ProtocolHandler
}

type AjpProtocolHandlerImpl struct {
	impl.ProtocolHandlerImpl
}

func NewAjpProtocolHandler(
	packetUnpacker protocol.PacketUnpacker,
	packetPacker protocol.PacketPacker,
	commandUnpacker protocol.CommandUnpacker,
	commandPacker protocol.CommandPacker,
	commandHandler protocol.CommandHandler,
	serverMode bool) *AjpProtocolHandlerImpl {

	ph := AjpProtocolHandlerImpl{}
	ph.ProtocolHandlerImpl.ConstructProtocolHandler(packetUnpacker, packetPacker, commandUnpacker, commandPacker, commandHandler, serverMode)
	return &ph
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *AjpProtocolHandlerImpl) Reset() {
	h.ProtocolHandlerImpl.Reset()
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (h *AjpProtocolHandlerImpl) Protocol() string {
	return AJP_PROTO_NAME
}

func (h *AjpProtocolHandlerImpl) MaxReqPacketDataSize() int {
	return AJP_MAX_DATA_LEN
}
func (h *AjpProtocolHandlerImpl) MaxResPacketDataSize() int {
	return AJP_MAX_CHUNKLEN
}
