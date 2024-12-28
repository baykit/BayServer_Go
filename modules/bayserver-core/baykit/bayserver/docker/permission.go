package docker

import (
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/tour"
)

type Permission interface {
	Docker

	SocketAdmitted(rd rudder.Rudder) exception.HttpException

	TourAdmitted(tur tour.Tour) exception.HttpException
}
