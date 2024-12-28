package tour

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

type TourHandler interface {

	// Send HTTP headers to client
	SendHeaders(tour Tour) exception2.IOException

	// Send Contents to client
	SendContent(tour Tour, bytes []byte, ofs int, length int, lis common.DataConsumeListener) exception2.IOException

	// Send end of contents to client.
	SendEnd(tour Tour, keepAlive bool, lis common.DataConsumeListener) exception2.IOException

	// Send protocol error to client
	OnProtocolError(e exception.ProtocolException) (bool, exception2.IOException)
}
