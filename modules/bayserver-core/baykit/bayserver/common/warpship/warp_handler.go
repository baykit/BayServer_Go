package warpship

import (
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/exception"
)

type WarpHandler interface {
	tour.TourHandler

	NextWarpId() int
	NewWarpData(warpId int) *WarpData

	/**
	 * Verifies if protocol is allowed
	 */
	VerifyProtocol(protocol string) exception.IOException
}
