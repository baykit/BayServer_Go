package h1

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-docker-http/baykit/bayserver/docker/http"
)

type H1ProtocolHandler interface {
	protocol.ProtocolHandler
}

type H1ProtocolHandlerImpl struct {
	impl.ProtocolHandlerImpl
	Keeping bool
}

func NewH1ProtocolHandler(
	packetUnpacker protocol.PacketUnpacker,
	packetPacker protocol.PacketPacker,
	commandUnpacker protocol.CommandUnpacker,
	commandPacker protocol.CommandPacker,
	commandHandler protocol.CommandHandler,
	serverMode bool) *H1ProtocolHandlerImpl {

	ph := H1ProtocolHandlerImpl{}
	ph.ProtocolHandlerImpl.ConstructProtocolHandler(packetUnpacker, packetPacker, commandUnpacker, commandPacker, commandHandler, serverMode)
	return &ph
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *H1ProtocolHandlerImpl) Reset() {
	h.ProtocolHandlerImpl.Reset()
	h.Keeping = false
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (h *H1ProtocolHandlerImpl) Protocol() string {
	return http.H1_PROTO_NAME
}

func (h *H1ProtocolHandlerImpl) MaxReqPacketDataSize() int {
	return MAX_DATA_LEN
}
func (h *H1ProtocolHandlerImpl) MaxResPacketDataSize() int {
	return MAX_DATA_LEN
}
