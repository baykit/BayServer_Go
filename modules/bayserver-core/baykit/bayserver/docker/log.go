package docker

import (
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/exception"
)

type Log interface {
	Docker

	Log(tur tour.Tour) exception.IOException
}
