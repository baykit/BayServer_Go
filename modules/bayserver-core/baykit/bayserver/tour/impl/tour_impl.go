package impl

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/inboundship"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"strconv"
)

const STATE_UNINITIALIZED = 1
const STATE_PREPARING = 2
const STATE_READING = 3
const STATE_RUNNING = 4
const STATE_ABORTED = 5
const STATE_ENDED = 6
const STATE_ZOMBIE = 7

const TOUR_ID_NOCHECK = -1
const INVALID_TOUR_ID = 0

var oidCounter = util.NewCounter()
var idCounter = util.NewCounter()

type TourImpl struct {
	ship     inboundship.InboundShip
	shipId   int
	objectId int

	tourId        int
	errorHandling bool
	town          docker.Town
	city          docker.City
	club          docker.Club

	req *TourReqImpl
	res *TourResImpl

	interval int
	isSecure bool
	state    int
	error    exception.HttpException
}

func NewTour() tour.Tour {
	tur := TourImpl{
		objectId: oidCounter.Next(),
		state:    STATE_UNINITIALIZED,
	}
	tur.req = NewTourReq(&tur)
	tur.res = NewTourRes(&tur)
	return &tur
}

func (tur *TourImpl) String() string {
	var shipStr = ""
	if tur.ship != nil {
		shipStr = tur.ship.String()
	}
	return shipStr + " tour#" + strconv.Itoa(tur.tourId) + "/" + strconv.Itoa(tur.objectId) + "[key=" + strconv.Itoa(tur.req.key) + "]"
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (tur *TourImpl) Reset() {
	baylog.Debug("%s Tour reset", tur)

	tur.req.Reset()
	tur.res.Reset()
	tur.city = nil
	tur.town = nil
	tur.club = nil
	tur.errorHandling = false

	tur.ChangeState(TOUR_ID_NOCHECK, STATE_UNINITIALIZED)
	tur.tourId = INVALID_TOUR_ID

	tur.interval = 0
	tur.isSecure = false
	tur.error = nil
	tur.ship = nil
	tur.shipId = ship.INVALID_SHIP_ID
}

/****************************************/
/* Implements Tour                      */
/****************************************/

func (tur *TourImpl) Init(key int, sip ship.Ship) {
	if tur.IsInitialized() {
		bayserver.FatalError(exception2.NewSink("%s Tour already initialized: state=%s", tur, tur.state))
	}

	tur.ship = sip.(inboundship.InboundShip)
	tur.shipId = sip.ShipId()
	tur.tourId = idCounter.Next()
	tur.Req().(*TourReqImpl).key = key

	tur.Req().Init(key)
	tur.Res().Init()

	tur.ChangeState(TOUR_ID_NOCHECK, STATE_PREPARING)
	baylog.Debug("%s Tour initialized", tur)
}

func (tur *TourImpl) TourId() int {
	return tur.tourId
}

func (tur *TourImpl) ShipId() int {
	return tur.shipId
}

func (tur *TourImpl) Req() tour.TourReq {
	return tur.req
}

func (tur *TourImpl) Res() tour.TourRes {
	return tur.res
}

func (tur *TourImpl) City() interface{} {
	return tur.city
}

func (tur *TourImpl) SetCity(city interface{}) {
	tur.city = city.(docker.City)
}

func (tur *TourImpl) Town() interface{} {
	return tur.town
}

func (tur *TourImpl) SetTown(town interface{}) {
	tur.town = town.(docker.Town)
}

func (tur *TourImpl) SetClub(club interface{}) {
	tur.club = club.(docker.Club)
}

func (tur *TourImpl) Ship() interface{} {
	return tur.ship
}

func (tur *TourImpl) State() int {
	return tur.state
}

func (tur *TourImpl) Secure() bool {
	return tur.isSecure
}

func (tur *TourImpl) SetSecure(secure bool) {
	tur.isSecure = secure
}

func (tur *TourImpl) SetErrorHandling(handling bool) {
	tur.errorHandling = true
}

func (tur *TourImpl) IsValid() bool {
	return tur.state == STATE_PREPARING || tur.state == STATE_READING || tur.state == STATE_RUNNING
}

func (tur *TourImpl) IsPreparing() bool {
	return tur.state == STATE_PREPARING
}

func (tur *TourImpl) IsReading() bool {
	return tur.state == STATE_READING
}

func (tur *TourImpl) IsRunning() bool {
	return tur.state == STATE_RUNNING
}

func (tur *TourImpl) IsAborted() bool {
	return tur.state == STATE_ABORTED
}

func (tur *TourImpl) IsZombie() bool {
	return tur.state == STATE_ZOMBIE
}

func (tur *TourImpl) IsEnded() bool {
	return tur.state == STATE_ENDED
}

func (tur *TourImpl) IsInitialized() bool {
	return tur.state != STATE_UNINITIALIZED
}

func (tur *TourImpl) SetHttpError(hterr exception.HttpException) {
	tur.error = hterr
}

func (tur *TourImpl) CheckTourId(checkId int) {
	if checkId == TOUR_ID_NOCHECK {
		return
	}

	if !tur.IsInitialized() {
		bayserver.FatalError(exception2.NewSink("%s Tour not initialized", tur))
	}
	if checkId != tur.tourId {
		bayserver.FatalError(exception2.NewSink("%s Invalid tour id : %d", tur, tur.tourId))
	}
}

func (tur *TourImpl) Go() exception.HttpException {

	tur.city = tur.ship.PortDocker().FindCity(tur.req.reqHost)
	if tur.city == nil {
		tur.city = bayserver.FindCity(tur.req.reqHost)
	}

	if tur.req.headers.ContentLength() > 0 {
		tur.ChangeState(TOUR_ID_NOCHECK, STATE_READING)

	} else {
		tur.ChangeState(TOUR_ID_NOCHECK, STATE_RUNNING)

	}

	baylog.Debug("%s GO TOUR! ...( ^_^)/: city=%s url=%s", tur, tur.req.reqHost, tur.req.uri)

	if tur.city == nil {
		return exception.NewHttpException(httpstatus.NOT_FOUND, tur.req.uri)

	} else {
		return tur.city.Enter(tur)
	}
}

func (tur *TourImpl) Interval() int {
	return tur.interval
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (tur *TourImpl) ChangeState(checkId int, newState int) {
	tur.CheckTourId(checkId)
	tur.state = newState
}

func (tur *TourImpl) Error() exception.HttpException {
	return tur.error
}

/****************************************/
/* Private functions                    */
/****************************************/
