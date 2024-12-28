package h2

import (
	"bayserver-core/baykit/bayserver/common/exception"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

type H2Handler interface {
	// extends H2CommandHandler

	/**
	 * Send protocol error to client
	 */
	OnProtocolError(e exception.ProtocolException) (bool, exception2.IOException)
}
