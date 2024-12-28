package h1

import (
	"bayserver-core/baykit/bayserver/common/exception"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

type H1Handler interface {
	H1CommandHandler

	/**
	 * Send protocol error to client
	 */
	OnProtocolError(e exception.ProtocolException) (bool, exception2.IOException)
}
