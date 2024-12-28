package ajp

import (
	"bayserver-core/baykit/bayserver/common/exception"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

type AjpHandler interface {
	AjpCommandHandler

	/**
	 * Send protocol error to client
	 */
	OnProtocolError(e exception.ProtocolException) (bool, exception2.IOException)
}
