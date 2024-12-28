package impl

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/warpship"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/protocol"
	"bayserver-core/baykit/bayserver/rudder"
	ship "bayserver-core/baykit/bayserver/ship/impl"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"strconv"
	"sync"
)

type WarpShipImpl struct {
	ship.ShipImpl

	docker           docker.Warp
	tourMap          map[int]*util.Pair[int, tour.Tour]
	tourMapLock      sync.Mutex
	protocolHandler  protocol.ProtocolHandler
	connected        bool
	socketTimeoutSec int
	cmdBuf           []*util.Pair[protocol.Command, common.DataConsumeListener]
}

func NewWarpShip() warpship.WarpShip {
	ret := WarpShipImpl{
		tourMap: make(map[int]*util.Pair[int, tour.Tour]),
		cmdBuf:  make([]*util.Pair[protocol.Command, common.DataConsumeListener], 0),
	}
	ret.ShipImpl.Construct()
	return &ret
}

func (sip *WarpShipImpl) String() string {
	protocol := ""
	if sip.protocolHandler != nil {
		protocol = "[" + sip.protocolHandler.Protocol() + "]"
	}
	return "agt#" + strconv.Itoa(sip.AgentId()) + " wsip#" + strconv.Itoa(sip.ShipId()) + "/" + strconv.Itoa(sip.ObjectId()) + protocol
}

func (sip *WarpShipImpl) InitWarp(
	rd rudder.Rudder,
	agtId int,
	tp common.Transporter,
	dkr docker.Warp,
	protoHandler protocol.ProtocolHandler) {

	sip.ShipImpl.Init(agtId, rd, tp)
	sip.docker = dkr
	if dkr.TimeoutSec() >= 0 {
		sip.socketTimeoutSec = dkr.TimeoutSec()
	} else {
		sip.socketTimeoutSec = bayserver.Harbor().SocketTimeoutSec()
	}
	sip.protocolHandler = protoHandler
	protoHandler.Init(sip)
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (sip *WarpShipImpl) Reset() {
	sip.ShipImpl.Reset()
	if len(sip.tourMap) > 0 {
		baylog.Error("BUG: Some tours is active: %s", sip.tourMap)
	}
	clear(sip.tourMap)
	sip.connected = false
	sip.cmdBuf = sip.cmdBuf[:0]
	sip.protocolHandler = nil
}

/****************************************/
/* Implements Ship                      */
/****************************************/

func (sip *WarpShipImpl) NotifyHandshakeDone(protocol string) (common.NextSocketAction, exception.IOException) {
	ioerr := sip.ProtocolHandler().(warpship.WarpHandler).VerifyProtocol(protocol)
	if ioerr != nil {
		return -1, ioerr
	}
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (sip *WarpShipImpl) NotifyConnect() (common.NextSocketAction, exception.IOException) {
	baylog.Debug("%s notifyConnect", sip)
	sip.connected = true
	for _, pir := range sip.tourMap {
		tur := pir.B
		tur.CheckTourId(pir.A)
		ioerr := warpship.WarpDataGet(tur).Start()
		if ioerr != nil {
			return -1, ioerr
		}
	}
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (sip *WarpShipImpl) NotifyRead(buf []byte) (common.NextSocketAction, exception.IOException) {
	return sip.protocolHandler.BytesReceived(buf)
}

func (sip *WarpShipImpl) NotifyEof() common.NextSocketAction {
	baylog.Debug("%s EOF detected", sip)

	if len(sip.tourMap) == 0 {
		baylog.Debug("%s No warp tour. only close", sip)
		return common.NEXT_SOCKET_ACTION_CLOSE
	}

	for _, pir := range sip.tourMap {
		tur := pir.B
		tur.CheckTourId(pir.A)

		var ioerr exception.IOException = nil
		if !tur.Res().HeaderSent() {
			baylog.Debug("%s Send ServiceUnavailable: tur=%s", sip, tur)
			ioerr = tur.Res().SendError(impl.TOUR_ID_NOCHECK, httpstatus.SERVICE_UNAVAILABLE, "Server closed on reading headers", nil)

		} else {
			// NOT treat EOF as Error
			baylog.Debug("%s EOF is not an error: tur=%s", sip, tur)
			ioerr = tur.Res().EndResContent(impl.TOUR_ID_NOCHECK)
		}

		if ioerr != nil {
			baylog.DebugE(ioerr, "")
		}
	}

	clear(sip.tourMap)

	return common.NEXT_SOCKET_ACTION_CLOSE
}

func (sip *WarpShipImpl) NotifyError(e exception.Exception) {
	baylog.DebugE(e, "%s Error notified", sip)
}

func (sip *WarpShipImpl) NotifyProtocolError(e exception2.ProtocolException) (bool, exception.IOException) {
	baylog.ErrorE(e, "")
	sip.notifyErrorToOwnerTour(httpstatus.SERVICE_UNAVAILABLE, e.Error())
	return true, nil
}

func (sip *WarpShipImpl) CheckTimeout(durationSec int) bool {
	if sip.isTimedOut(durationSec) {
		sip.notifyErrorToOwnerTour(httpstatus.SERVICE_UNAVAILABLE, sip.String()+" server timed out")
		return true

	} else {
		return false
	}
}

func (sip *WarpShipImpl) NotifyClose() {
	baylog.Debug("%s notifyClose", sip)
	sip.notifyErrorToOwnerTour(httpstatus.SERVICE_UNAVAILABLE, sip.String()+" server closed")
	sip.endShip()
}

/****************************************/
/* Implements WarpShip                  */
/****************************************/

func (sip *WarpShipImpl) ProtocolHandler() protocol.ProtocolHandler {
	return sip.protocolHandler
}

func (sip *WarpShipImpl) Docker() docker.Warp {
	return sip.docker
}

func (sip *WarpShipImpl) WarpHandler() warpship.WarpHandler {
	return sip.protocolHandler.CommandHandler().(warpship.WarpHandler)
}

func (sip *WarpShipImpl) Abort(chkId int) {
	sip.CheckShipId(chkId)
	sip.Transporter().ReqClose(sip.Rudder())
}

func (sip *WarpShipImpl) Flush() exception.IOException {
	baylog.Debug("%s flush", sip)
	for _, cmdAndLis := range sip.cmdBuf {
		if cmdAndLis.A == nil {
			cmdAndLis.B()

		} else {
			ioerr := sip.protocolHandler.Post(cmdAndLis.A, cmdAndLis.B)
			if ioerr != nil {
				return ioerr
			}
		}
	}
	sip.cmdBuf = sip.cmdBuf[:0]
	return nil
}

func (sip *WarpShipImpl) GetTour(warpId int, must bool) (tour.Tour, exception2.ProtocolException) {
	pir := sip.tourMap[warpId]
	if pir != nil {
		tur := pir.B
		tur.CheckTourId(pir.A)
		if !warpship.WarpDataGet(tur).Ended {
			return tur, nil
		}
	}

	if must {
		return nil, exception2.NewProtocolException("%s warp tour not found: id=%d", sip, warpId)

	} else {
		return nil, nil
	}
}

func (sip *WarpShipImpl) Initialized() bool {
	return sip.ShipImpl.Initialized
}

func (sip *WarpShipImpl) StartWarpTour(tur tour.Tour) exception.IOException {

	var ioerr exception.IOException = nil

	for { // try catch
		wHnd := sip.warpHandler()
		warpId := wHnd.NextWarpId()
		wdat := wHnd.NewWarpData(warpId)
		baylog.Debug("%s new warp tour related to %s", wdat, tur)
		tur.Req().SetReqContentHandler(wdat)

		baylog.Debug("%s start: warpId=%d", wdat, warpId)
		if _, exists := sip.tourMap[warpId]; exists {
			bayserver.FatalError(exception.NewSink("warpId exists"))
		}

		sip.tourMap[warpId] = util.NewPair(tur.TourId(), tur)
		ioerr = wHnd.SendHeaders(tur)
		if ioerr != nil {
			break
		}

		if sip.connected {
			baylog.Debug("%s is already connected. Start warp tour:%s", wdat, tur)
			ioerr = wdat.Start()
			if ioerr != nil {
				break
			}
		}

		return nil
	}

	return ioerr
}

func (sip *WarpShipImpl) EndWarpTour(tur tour.Tour, keep bool) {
	wdat := warpship.WarpDataGet(tur)
	baylog.Debug("%s %s end: started=%t ended=%t keep=%t", sip, tur, wdat.Started, wdat.Ended, keep)
	delete(sip.tourMap, wdat.WarpId)
	if keep {
		baylog.Debug("%s keep warp ship", sip)
		sip.docker.Keep(sip)
	}
}

func (sip *WarpShipImpl) Post(cmd protocol.Command, lis common.DataConsumeListener) exception.IOException {
	baylog.Debug("%s post: cmd=%s", sip, cmd)
	if !sip.connected {
		p := util.NewPair(cmd, lis)
		sip.cmdBuf = append(sip.cmdBuf, p)

	} else {
		if cmd == nil {
			lis()
		} else {
			ioerr := sip.protocolHandler.Post(cmd, lis)
			if ioerr != nil {
				return ioerr
			}
		}
	}

	return nil
}

/****************************************/
/* Private functions                    */
/****************************************/
func (sip *WarpShipImpl) warpHandler() warpship.WarpHandler {
	return sip.protocolHandler.CommandHandler().(warpship.WarpHandler)
}

func (sip *WarpShipImpl) notifyErrorToOwnerTour(status int, msg string) {
	sip.tourMapLock.Lock()
	defer sip.tourMapLock.Unlock()

	for _, pir := range sip.tourMap {
		tur := pir.B
		baylog.Debug("%s send error to owner: %s running=%t", sip, tur, tur.IsRunning())

		var ioerr exception.IOException = nil
		if tur.IsRunning() || tur.IsReading() {
			ioerr = tur.Res().SendError(impl.TOUR_ID_NOCHECK, status, msg, nil)

		} else {
			ioerr = tur.Res().EndResContent(impl.TOUR_ID_NOCHECK)
		}

		if ioerr != nil {
			baylog.ErrorE(ioerr, "")
		}
	}

	clear(sip.tourMap)
}

func (sip *WarpShipImpl) endShip() {
	sip.docker.OnEndShip(sip)
}

func (sip *WarpShipImpl) isTimedOut(durationSec int) bool {
	var timedOut bool
	if sip.Keeping {
		// warp connection never timeout in keeping
		timedOut = false

	} else if sip.socketTimeoutSec <= 0 {
		timedOut = false

	} else {
		timedOut = durationSec >= sip.socketTimeoutSec
	}

	baylog.Debug("%s Warp check timeout: dur=%d, timeout=%t, keeping=%t limit=%d", sip, durationSec, timedOut, sip.Keeping, sip.socketTimeoutSec)
	return timedOut
}
