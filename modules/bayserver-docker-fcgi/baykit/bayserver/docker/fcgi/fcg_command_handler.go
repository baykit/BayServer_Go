package fcgi

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

type FcgCommandHandler interface {
	protocol.CommandHandler
	HandleBeginRequest(cmd *CmdBeginRequest) (common.NextSocketAction, exception.IOException)
	HandleEndRequest(cmd *CmdEndRequest) (common.NextSocketAction, exception.IOException)
	HandleParams(cmd *CmdParams) (common.NextSocketAction, exception.IOException)
	HandleStdErr(cmd *CmdStdErr) (common.NextSocketAction, exception.IOException)
	HandleStdIn(cmd *CmdStdIn) (common.NextSocketAction, exception.IOException)
	HandleStdOut(cmd *CmdStdOut) (common.NextSocketAction, exception.IOException)
}
