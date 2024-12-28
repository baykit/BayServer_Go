package h2

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/impl"
	"bayserver-docker-http/baykit/bayserver/docker/http"
)

const CTL_STREAM_ID = 0

type H2ProtocolHandler interface {
	protocol.ProtocolHandler
}

type H2ProtocolHandlerImpl struct {
	impl.ProtocolHandlerImpl
	Keeping bool
}

func NewH2ProtocolHandler(
	packetUnpacker protocol.PacketUnpacker,
	packetPacker protocol.PacketPacker,
	commandUnpacker protocol.CommandUnpacker,
	commandPacker protocol.CommandPacker,
	commandHandler protocol.CommandHandler,
	serverMode bool) *H2ProtocolHandlerImpl {

	ph := H2ProtocolHandlerImpl{}
	ph.ProtocolHandlerImpl.ConstructProtocolHandler(packetUnpacker, packetPacker, commandUnpacker, commandPacker, commandHandler, serverMode)
	return &ph
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *H2ProtocolHandlerImpl) Reset() {
	h.ProtocolHandlerImpl.Reset()
	h.Keeping = false
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (h *H2ProtocolHandlerImpl) Protocol() string {
	return http.H2_PROTO_NAME
}

func (h *H2ProtocolHandlerImpl) MaxReqPacketDataSize() int {
	return DEFAULT_PAYLOAD_MAXLEN
}
func (h *H2ProtocolHandlerImpl) MaxResPacketDataSize() int {
	return DEFAULT_PAYLOAD_MAXLEN
}
