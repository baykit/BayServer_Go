package protocol

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/exception"
)

type CommandUnpacker interface {
	util.Reusable
	PacketReceived(pkt Packet) (common.NextSocketAction, exception.IOException)
}
