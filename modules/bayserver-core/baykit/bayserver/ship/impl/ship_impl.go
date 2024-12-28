package ship

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"strconv"
)

var oidCounter = util.NewCounter()
var idCounter = util.NewCounter()

type ShipImpl struct {
	// util.Reusable implements
	objectId        int
	shipId          int
	agentId         int
	rudder          rudder.Rudder
	transporter     common.Transporter
	Ch              int
	ProtocolHandler protocol.ProtocolHandler
	Initialized     bool
	Keeping         bool
}

func (s *ShipImpl) Construct() {
	s.objectId = oidCounter.Next()
	s.shipId = ship.INVALID_SHIP_ID
}

func (s *ShipImpl) Init(
	agentId int,
	rd rudder.Rudder,
	tp common.Transporter) {

	if s.Initialized {
		baylog.FatalE(exception.NewSink("Ship already initialized"), "")
		panic("Ship already initialized")
	}
	s.shipId = idCounter.Next()
	s.agentId = agentId
	s.rudder = rd
	s.transporter = tp
	s.Initialized = true
}

func (s *ShipImpl) String() string {
	return "ship#" + strconv.Itoa(s.shipId)
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (s *ShipImpl) Reset() {
	baylog.Debug("%s reset", s)
	s.Initialized = false
	s.rudder = nil
	s.transporter = nil
	s.agentId = -1
	s.shipId = ship.INVALID_SHIP_ID
	s.Keeping = false
}

/****************************************/
/* Implements Ship                      */
/****************************************/

func (s *ShipImpl) ObjectId() int {
	return s.objectId
}

func (s *ShipImpl) ShipId() int {
	return s.shipId
}

func (s *ShipImpl) AgentId() int {
	return s.agentId
}

func (s *ShipImpl) Rudder() rudder.Rudder {
	return s.rudder
}

func (s *ShipImpl) Transporter() common.Transporter {
	return s.transporter
}

func (s *ShipImpl) CheckShipId(checkId int) {
	if !s.Initialized {
		bayserver.FatalError(exception.NewSink("%s Uninitialized ship (might be returned ship): %d", s, checkId))
	}
	if checkId == 0 || (checkId != ship.SHIP_ID_NOCHECK && checkId != s.shipId) {
		bayserver.FatalError(exception.NewSink("%s Invalid ship id (might be returned ship): %d", s, checkId))
	}
}

func (s *ShipImpl) ResumeRead(checkId int) {
	s.CheckShipId(checkId)
	baylog.Debug("%s resume read", s)
	s.transporter.ReqRead(s.rudder)

}

func (s *ShipImpl) PostClose(checkId int) {
	s.CheckShipId(checkId)
	s.transporter.ReqClose(s.rudder)
}
