package h1

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/util/exception"
)

type H1CommandHandler interface {
	protocol.CommandHandler
	HandleHeader(cmd *CmdHeader) (common.NextSocketAction, exception.IOException)
	HandleContent(cmd *CmdContent) (common.NextSocketAction, exception.IOException)
	HandleEndContent(cmd *CmdEndContent) (common.NextSocketAction, exception.IOException)
	ReqFinished() bool
}
