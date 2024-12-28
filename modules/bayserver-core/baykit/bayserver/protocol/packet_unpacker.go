package protocol

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/exception"
)

type PacketUnpacker interface {
	util.Reusable
	BytesReceived(bytes []byte) (common.NextSocketAction, exception.IOException)
}
