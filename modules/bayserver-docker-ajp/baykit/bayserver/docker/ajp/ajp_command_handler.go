package ajp

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

type AjpCommandHandler interface {
	protocol.CommandHandler
	HandleData(cmd *CmdData) (common.NextSocketAction, exception.IOException)
	HandleEndResponse(cmd *CmdEndResponse) (common.NextSocketAction, exception.IOException)
	HandleForwardRequest(cmd *CmdForwardRequest) (common.NextSocketAction, exception.IOException)
	HandleSendBodyChunk(cmd *CmdSendBodyChunk) (common.NextSocketAction, exception.IOException)
	HandleSendHeaders(cmd *CmdSendHeaders) (common.NextSocketAction, exception.IOException)
	HandleShutdown(cmd *CmdShutdown) (common.NextSocketAction, exception.IOException)
	HandleGetBodyChunk(cmd *CmdGetBodyChunk) (common.NextSocketAction, exception.IOException)
	NeedData() bool
}
