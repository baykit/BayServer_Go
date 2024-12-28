package protocol

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/exception"
)

type ProtocolHandler interface {
	util.Reusable
	PacketUnpacker() PacketUnpacker
	PacketPacker() PacketPacker
	CommandUnpacker() CommandUnpacker
	CommandPacker() CommandPacker
	CommandHandler() CommandHandler
	ServerMode() bool
	Ship() ship.Ship
	Protocol() string

	Init(ship ship.Ship)
	BytesReceived(buf []byte) (common.NextSocketAction, exception.IOException)
	Post(command Command, listener common.DataConsumeListener) exception.IOException

	MaxReqPacketDataSize() int
	MaxResPacketDataSize() int
}
