package protocol

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/exception"
)

type PacketPacker interface {
	util.Reusable
	Post(sip ship.Ship, pkt Packet, lis common.DataConsumeListener) exception.IOException
}
