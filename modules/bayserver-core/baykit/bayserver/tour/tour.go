package tour

import (
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util"
)

type Tour interface {
	util.Reusable
	String() string
	Init(key int, ship ship.Ship)

	TourId() int
	ShipId() int
	Req() TourReq
	Res() TourRes

	City() interface{}        // docker.City
	SetCity(city interface{}) // docker.City
	Town() interface{}
	SetTown(town interface{}) // docker.Town
	SetClub(club interface{}) // docker.Club
	State() int
	Ship() interface{}

	Secure() bool
	SetSecure(bool)

	IsValid() bool
	IsPreparing() bool
	IsReading() bool
	IsRunning() bool
	IsAborted() bool
	IsZombie() bool
	IsEnded() bool
	IsInitialized() bool

	SetHttpError(hterr exception.HttpException)

	CheckTourId(checkId int)

	Go() exception.HttpException
	Interval() int
	Error() exception.HttpException
	SetErrorHandling(handling bool)
}
