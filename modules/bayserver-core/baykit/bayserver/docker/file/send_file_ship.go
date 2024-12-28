package file

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/ship"
	impl "bayserver-core/baykit/bayserver/ship/impl"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"strconv"
)

type SendFileShip struct {
	impl.ShipImpl
	fileWroteLen int
	fileSize     int
	tour         tour.Tour
	tourId       int
}

func NewSendFileShip() *SendFileShip {
	res := SendFileShip{}
	res.ShipImpl.Construct()
	return &res
}

func (s *SendFileShip) String() string {
	return "agt#" + strconv.Itoa(s.AgentId()) + " file#" + strconv.Itoa(s.ShipId()) + "/" + strconv.Itoa(s.ObjectId())
}

func (s *SendFileShip) Init(rd rudder.Rudder, tp common.Transporter, tur tour.Tour, fileSize int) {
	s.ShipImpl.Init(tur.Ship().(ship.Ship).AgentId(), rd, tp)
	s.tour = tur
	s.tourId = tur.TourId()
	s.fileSize = fileSize
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (s *SendFileShip) Reset() {
	s.ShipImpl.Reset()
	s.fileWroteLen = 0
	s.fileSize = 0
	s.tour = nil
	s.tourId = 0
}

/****************************************/
/* Implements Ship                      */
/****************************************/

func (s *SendFileShip) NotifyHandshakeDone(protocol string) (common.NextSocketAction, exception2.IOException) {
	bayserver.FatalError(exception2.NewSink("Not implemented"))
	return 0, nil
}

func (s *SendFileShip) NotifyConnect() (common.NextSocketAction, exception2.IOException) {
	bayserver.FatalError(exception2.NewSink("Not implemented"))
	return 0, nil
}

func (s *SendFileShip) NotifyRead(buf []byte) (common.NextSocketAction, exception2.IOException) {
	s.fileWroteLen += len(buf)
	baylog.Debug("%s read file %d bytes: total=%d/%d", s, len(buf), s.fileWroteLen, s.fileSize)
	available, ioerr := s.tour.Res().SendResContent(s.tourId, buf, 0, len(buf))

	if ioerr != nil {
		return 0, ioerr
	}

	if s.fileWroteLen >= s.fileSize {
		s.NotifyEof()
		return common.NEXT_SOCKET_ACTION_CLOSE, nil
	} else if available {
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	} else {
		return common.NEXT_SOCKET_ACTION_SUSPEND, nil
	}
}

func (s *SendFileShip) NotifyError(e exception2.Exception) {
	baylog.DebugE(e, "%s Notify Error", s)

	ioerr := s.tour.Res().SendError(s.tourId, httpstatus.INTERNAL_SERVER_ERROR, "", e)

	if ioerr != nil {
		baylog.DebugE(ioerr, "")
	}
}

func (s *SendFileShip) NotifyEof() common.NextSocketAction {
	baylog.Debug("%s EOF", s)
	ioerr := s.tour.Res().EndResContent(s.tourId)
	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
	}
	return common.NEXT_SOCKET_ACTION_CLOSE
}

func (s *SendFileShip) NotifyProtocolError(e exception.ProtocolException) (bool, exception2.IOException) {
	bayserver.FatalError(exception2.NewSink("Not implemented"))
	return false, nil
}

func (s *SendFileShip) NotifyClose() {

}

func (s *SendFileShip) CheckTimeout(durationSec int) bool {
	return false
}
