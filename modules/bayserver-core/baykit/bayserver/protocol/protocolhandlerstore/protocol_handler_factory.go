package protocolhandlerstore

import (
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
)

type ProtocolHandlerFactory func(pktStore *packetstore.PacketStore) protocol.ProtocolHandler
