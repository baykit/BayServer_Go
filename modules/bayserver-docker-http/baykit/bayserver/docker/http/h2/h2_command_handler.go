package h2

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

type H2CommandHandler interface {
	protocol.CommandHandler
	HandlePreface(cmd *CmdPreface) (common.NextSocketAction, exception.IOException)
	HandleData(cmd *CmdData) (common.NextSocketAction, exception.IOException)
	HandleHeaders(cmd *CmdHeaders) (common.NextSocketAction, exception.IOException)
	HandlePriority(cmd *CmdPriority) (common.NextSocketAction, exception.IOException)
	HandleSettings(cmd *CmdSettings) (common.NextSocketAction, exception.IOException)
	HandleWindowUpdate(cmd *CmdWindowUpdate) (common.NextSocketAction, exception.IOException)
	HandleGoAway(cmd *CmdGoAway) (common.NextSocketAction, exception.IOException)
	HandlePing(cmd *CmdPing) (common.NextSocketAction, exception.IOException)
	HandleRstStream(cmd *CmdRstStream) (common.NextSocketAction, exception.IOException)
}
