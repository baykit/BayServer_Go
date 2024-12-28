package protocol

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/exception"
)

type CommandPacker interface {
	util.Reusable
	Post(sip ship.Ship, c Command, listener common.DataConsumeListener) exception.IOException
}
