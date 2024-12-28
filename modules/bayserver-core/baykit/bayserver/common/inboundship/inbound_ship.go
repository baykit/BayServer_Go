package inboundship

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/exception"
)

type InboundShip interface {
	ship.Ship

	PortDocker() docker.Port

	ProtocolHandler() protocol.ProtocolHandler
	SetProtocolHandler(protoHandler protocol.ProtocolHandler)
	GetTour(turKey int, force bool, rent bool) tour.Tour
	GetErrorTour() tour.Tour

	SendHeaders(checkId int, tour tour.Tour) exception.IOException
	SendResContent(checkId int, tour tour.Tour, bytes []byte, ofs int, length int, lis common.DataConsumeListener) exception.IOException
	SendEndTour(checkId int, tour tour.Tour, lis func()) exception.IOException
	ReturnTour(tour tour.Tour)
}
