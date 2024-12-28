package common

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/inboundship"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	shipimpl "bayserver-core/baykit/bayserver/ship/impl"
	"bayserver-core/baykit/bayserver/tour"
	tourimpl "bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/tour/tourstore"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"net"
	"strconv"
	"sync"
)

type InboundShipImpl struct {
	shipimpl.ShipImpl
	portDocker       docker.Port
	errorCounter     *util.Counter
	Conn             net.Conn
	protocolHandler  protocol.ProtocolHandler
	NeedEnd          bool
	SocketTimeoutSec int
	tourStore        *tourstore.TourStore
	activeTours      []tour.Tour
	lock             sync.Mutex
}

func NewInboundShip() inboundship.InboundShip {
	ret := InboundShipImpl{
		errorCounter: util.NewCounter(),
		activeTours:  []tour.Tour{},
	}
	ret.ShipImpl.Construct()
	return &ret
}

func (sip *InboundShipImpl) InitInbound(
	rd rudder.Rudder,
	agentId int,
	tp common.Transporter,
	portDkr docker.Port,
	protoHandler protocol.ProtocolHandler) {

	sip.Init(agentId, rd, tp)
	//baylog.Debug("%s InitInbound rd=%s", sip, rd)

	sip.Conn = impl.GetConn(rd)
	sip.portDocker = portDkr
	if portDkr.TimeoutSec() >= 0 {
		sip.SocketTimeoutSec = portDkr.TimeoutSec()
	} else {
		sip.SocketTimeoutSec = bayserver.Harbor().SocketTimeoutSec()
	}
	sip.tourStore = tourstore.GetStore(agentId)
	sip.SetProtocolHandler(protoHandler)
	baylog.Debug("%sip Initialized", sip)
}

func (sip *InboundShipImpl) String() string {
	proto := ""
	if sip.protocolHandler != nil {
		proto = "[" + sip.protocolHandler.Protocol() + "]"
	}
	return "agt#" + strconv.Itoa(sip.AgentId()) + " isip#" + strconv.Itoa(sip.ShipId()) + "/" + strconv.Itoa(sip.ObjectId()) + proto
}

func GetInboundShipImpl(sip inboundship.InboundShip) *InboundShipImpl {
	return sip.(*InboundShipImpl)
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (sip *InboundShipImpl) Reset() {
	sip.ShipImpl.Reset()
	sip.NeedEnd = false
	sip.Conn = nil
	sip.protocolHandler = nil
}

/****************************************/
/* Implements Ship                      */
/****************************************/

func (sip *InboundShipImpl) NotifyHandshakeDone(protocol string) (common.NextSocketAction, exception2.IOException) {
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (sip *InboundShipImpl) NotifyConnect() (common.NextSocketAction, exception2.IOException) {
	baylog.FatalE(exception2.NewSink("Illegal state"), "")
	panic("ERR")
}

func (sip *InboundShipImpl) NotifyRead(buf []byte) (common.NextSocketAction, exception2.IOException) {
	return sip.protocolHandler.BytesReceived(buf)
}

func (sip *InboundShipImpl) NotifyEof() common.NextSocketAction {
	baylog.Debug("%sip EOF detected", sip)
	return common.NEXT_SOCKET_ACTION_CLOSE
}
func (sip *InboundShipImpl) NotifyError(e exception2.Exception) {
	baylog.ErrorE(e, "")
}

func (sip *InboundShipImpl) NotifyProtocolError(e exception.ProtocolException) (bool, exception2.IOException) {
	baylog.DebugE(e, "")
	return sip.tourHandler().OnProtocolError(e)
}

func (sip *InboundShipImpl) NotifyClose() {
	baylog.Debug("%sip notifyClose", sip)

	sip.abortTours()

	if len(sip.activeTours) > 0 {
		// cannot close because there are some running tours
		baylog.Debug("%s cannot end ship because there are some running tours (ignore)", sip)
		sip.NeedEnd = true

	} else {
		sip.endShip()
	}
}

func (sip *InboundShipImpl) CheckTimeout(durationSec int) bool {
	var timeout bool
	if sip.SocketTimeoutSec <= 0 {
		timeout = false

	} else if sip.Keeping {
		timeout = durationSec >= bayserver.Harbor().KeepTimeoutSec()

	} else {
		timeout = durationSec >= sip.SocketTimeoutSec
	}

	baylog.Debug("%s Check timeout: dur=%d, timeout=%v, keeping=%v limit=%d keeplim=%d",
		sip, durationSec, timeout, sip.Keeping, sip.SocketTimeoutSec, bayserver.Harbor().KeepTimeoutSec())
	return timeout
}

/****************************************/
/* Implements InboundShip               */
/****************************************/

func (sip *InboundShipImpl) PortDocker() docker.Port {
	return sip.portDocker
}

func (sip *InboundShipImpl) ProtocolHandler() protocol.ProtocolHandler {
	return sip.protocolHandler
}

func (sip *InboundShipImpl) SetProtocolHandler(protoHandler protocol.ProtocolHandler) {
	sip.protocolHandler = protoHandler
	protoHandler.Init(sip)
	baylog.Debug("%sip protocol handler is set", sip)
}

func (sip *InboundShipImpl) GetTour(turKey int, force bool, rent bool) tour.Tour {
	sip.lock.Lock()
	defer sip.lock.Unlock()

	var tur tour.Tour = nil
	for { // try catch
		storeKey := sip.uniqKey(sip.ShipId(), turKey)
		//baylog.Debug("%s get tour key=%d uniqKey=%d", sip, turKey, storeKey)
		tur = sip.tourStore.Get(storeKey)
		//baylog.Debug("%s got tour from tour protocolhandlerstore:=%s", sip, tur)
		if tur == nil && rent {
			tur = sip.tourStore.Rent(storeKey, force)
			if tur == nil {
				break // return nil
			}
			tur.Init(turKey, sip)
			sip.activeTours = append(sip.activeTours, tur)
		}
		tur.CheckTourId(tur.TourId())
		break
	}
	return tur
}

func (sip *InboundShipImpl) GetErrorTour() tour.Tour {
	turKey := sip.errorCounter.Next()
	storeKey := sip.uniqKey(sip.ShipId(), -turKey)
	tur := sip.tourStore.Rent(storeKey, true)
	tur.Init(-turKey, sip)
	tur.SetErrorHandling(true)
	sip.activeTours = append(sip.activeTours, tur)
	return tur
}

func (sip *InboundShipImpl) SendHeaders(checkId int, tour tour.Tour) exception2.IOException {
	sip.CheckShipId(checkId)

	for _, nv := range sip.portDocker.AdditionalHeaders() {
		tour.Res().Headers().Add(nv[0], nv[1])
	}

	ioerr := sip.tourHandler().SendHeaders(tour)
	if ioerr != nil {
		return ioerr
	}
	return nil
}

func (sip *InboundShipImpl) SendResContent(checkId int, tour tour.Tour, bytes []byte, ofs int, length int, lis common.DataConsumeListener) exception2.IOException {
	sip.CheckShipId(checkId)

	maxLen := sip.protocolHandler.MaxResPacketDataSize()
	var ioerr exception2.IOException = nil
	for { // try/catch
		if length > maxLen {
			ioerr = sip.SendResContent(tourimpl.TOUR_ID_NOCHECK, tour, bytes, ofs, maxLen, nil)
			if ioerr != nil {
				break
			}
			ioerr = sip.SendResContent(tourimpl.TOUR_ID_NOCHECK, tour, bytes, ofs+maxLen, length-maxLen, lis)

		} else {
			ioerr = sip.tourHandler().SendContent(tour, bytes, ofs, length, lis)
		}

		break
	}

	if ioerr != nil {
		return ioerr
	}

	return nil
}

func (sip *InboundShipImpl) SendEndTour(checkId int, tur tour.Tour, lis func()) exception2.IOException {
	sip.CheckShipId(checkId)

	baylog.Debug("%s sendEndTour: %s state=%d", sip, tur, tur.State())

	if !tur.IsValid() {
		bayserver.FatalError(exception2.NewSink("Tour is not valid: state=%d", tur.State()))
	}

	keepAlive := false
	if tur.Req().Headers().GetConnection() == headers.CONNECTION_TYPE_KEEP_ALIVE {
		keepAlive = true
	}

	if keepAlive {
		resConn := tur.Res().Headers().GetConnection()
		keepAlive = (resConn == headers.CONNECTION_TYPE_KEEP_ALIVE) ||
			(resConn == headers.CONNECTION_TYPE_UNKNOWN)
		if keepAlive {
			clen := tur.Res().Headers().ContentLength()
			if clen < 0 {
				keepAlive = false
			}
		}
	}

	return sip.tourHandler().SendEnd(tur, keepAlive, lis)
}

func (sip *InboundShipImpl) ReturnTour(tur tour.Tour) {

	baylog.Debug("%s Return tour: %s", sip, tur)

	sip.lock.Lock()
	defer sip.lock.Unlock()

	if !arrayutil.Contains(sip.activeTours, tur) {
		bayserver.FatalError(exception2.NewSink("Tour is not in active list: %s", tur))
	}

	sip.tourStore.Return(sip.uniqKey(sip.ShipId(), tur.Req().Key()))
	sip.activeTours, _ = arrayutil.RemoveObject(sip.activeTours, tur)

	if sip.NeedEnd && len(sip.activeTours) == 0 {
		sip.endShip()
	}
}

/****************************************/
/* Private functions                    */
/****************************************/

func (sip *InboundShipImpl) tourHandler() tour.TourHandler {
	return sip.protocolHandler.CommandHandler().(tour.TourHandler)
}

func (sip *InboundShipImpl) uniqKey(sipId int, turKey int) int64 {
	return int64(sipId)<<32 | (int64(turKey) & 0xffff)
}

func (sip *InboundShipImpl) endShip() {
	baylog.Debug("%s endShip", sip)
	sip.portDocker.ReturnProtocolHandler(sip.AgentId(), sip.protocolHandler)
	sip.portDocker.ReturnShip(sip)
}

func (sip *InboundShipImpl) abortTours() {
	baylog.Debug("%s abort tours", sip)

	returnList := make([]tour.Tour, 0)

	// Abort tours
	for _, tur := range sip.activeTours {
		if tur.IsValid() {
			baylog.Debug("%s is valid, abort it: stat=%s", tur, tur.State())
			if tur.Req().(*tourimpl.TourReqImpl).Abort() {
				returnList = append(returnList, tur)
			}
		}
	}

	for _, tur := range returnList {
		sip.ReturnTour(tur)
	}
}
