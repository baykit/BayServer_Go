package warpship

import (
	"bayserver-core/baykit/bayserver/common"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/exception"
	"fmt"
)

type WarpShip interface {
	ship.Ship
	fmt.Stringer

	InitWarp(rd rudder.Rudder, agtId int, tp common.Transporter, dkr docker.Warp, protoHandler protocol.ProtocolHandler)
	Reset()
	ProtocolHandler() protocol.ProtocolHandler
	Abort(chkId int)
	Initialized() bool

	EndWarpTour(tur tour.Tour, keep bool)
	GetTour(warpId int, must bool) (tour.Tour, exception2.ProtocolException)
	Docker() docker.Warp
	WarpHandler() WarpHandler
	StartWarpTour(tur tour.Tour) exception.IOException
	Post(cmd protocol.Command, lis common.DataConsumeListener) exception.IOException
	Flush() exception.IOException
}
