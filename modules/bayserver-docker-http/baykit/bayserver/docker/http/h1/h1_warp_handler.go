package h1

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/warpship"
	"bayserver-core/baykit/bayserver/common/warpship/impl"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/tour"
	tourimpl "bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"strconv"
	"strings"
)

const STATE_READ_HEADER = 1
const STATE_READ_CONTENT = 2
const STATE_FINISHED = 3

const FIXED_WARP_ID = 1

type H1WarpHandler struct {
	protocolHandler *H1ProtocolHandlerImpl
	state           int
}

func NewH1WarpHandler() *H1WarpHandler {
	h := &H1WarpHandler{}
	h.resetState()

	var _ tour.TourHandler = h     // cast check
	var _ warpship.WarpHandler = h // cast check
	var _ H1Handler = h            // cast check
	var _ util.Reusable = h        // cast check
	return h
}

func (h *H1WarpHandler) Init(handler *H1ProtocolHandlerImpl) {
	h.protocolHandler = handler
}

func (h *H1WarpHandler) String() string {
	return "H1WarpHandler"
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *H1WarpHandler) Reset() {
	h.resetState()
}

/****************************************/
/* Implements TourHandler               */
/****************************************/

func (h *H1WarpHandler) SendHeaders(tur tour.Tour) exception2.IOException {

	town := tur.Town()

	townPath := town.(docker.Town).Name()
	if !strings.HasSuffix(townPath, "/") {
		townPath += "/"
	}

	sip := h.Ship()
	newUri := sip.Docker().DestTown() + tur.Req().Uri()[len(townPath):]

	cmd := NewReqHeader(tur.Req().Method(), newUri, "HTTP/1.1")

	for _, name := range tur.Req().Headers().HeaderNames() {
		for _, value := range tur.Req().Headers().HeaderValues(name) {
			cmd.AddHeader(name, value)
		}
	}

	if tur.Req().Headers().Contains(headers.X_FORWARDED_FOR) {
		cmd.SetHeader(headers.X_FORWARDED_FOR, tur.Req().Headers().Get(headers.X_FORWARDED_FOR))
	} else {
		cmd.SetHeader(headers.X_FORWARDED_FOR, tur.Req().RemoteAddress())
	}

	if tur.Req().Headers().Contains(headers.X_FORWARDED_PROTO) {
		cmd.SetHeader(headers.X_FORWARDED_PROTO, tur.Req().Headers().Get(headers.X_FORWARDED_PROTO))
	} else {
		proto := ""
		if tur.Secure() {
			proto = "https"
		} else {
			proto = "http"
		}
		cmd.SetHeader(headers.X_FORWARDED_PROTO, proto)
	}

	if tur.Req().Headers().Contains(headers.X_FORWARDED_PORT) {
		cmd.SetHeader(headers.X_FORWARDED_PORT, tur.Req().Headers().Get(headers.X_FORWARDED_PORT))
	} else {
		cmd.SetHeader(headers.X_FORWARDED_PORT, strconv.Itoa(tur.Req().ServerPort()))
	}

	if tur.Req().Headers().Contains(headers.X_FORWARDED_HOST) {
		cmd.SetHeader(headers.X_FORWARDED_HOST, tur.Req().Headers().Get(headers.X_FORWARDED_HOST))
	} else {
		cmd.SetHeader(headers.X_FORWARDED_HOST, tur.Req().Headers().Get(headers.HOST))
	}

	cmd.SetHeader(headers.HOST, sip.Docker().Host()+":"+strconv.Itoa(sip.Docker().Port()))
	cmd.SetHeader(headers.CONNECTION, "Keep-Alive")

	if bayserver.Harbor().TraceHeader() {
		for _, kv := range cmd.headers {
			baylog.Info("%s warp_http reqHdr: %s=%s", tur, kv[0], kv[1])
		}
	}

	return sip.Post(cmd, nil)
}

func (h *H1WarpHandler) SendContent(tur tour.Tour, buf []byte, start int, length int, lis common.DataConsumeListener) exception2.IOException {
	cmd := NewCmdContent(buf, start, length)
	return h.Ship().Post(cmd, lis)
}

func (h *H1WarpHandler) SendEnd(tur tour.Tour, keepAlive bool, lis common.DataConsumeListener) exception2.IOException {
	cmd := NewCmdEndContent()
	return h.Ship().Post(cmd, lis)
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (h *H1WarpHandler) OnProtocolError(e exception.ProtocolException) (bool, exception2.IOException) {
	bayserver.FatalError(exception2.NewSink(""))
	return false, nil
}

/****************************************/
/* Implements WarpHandler               */
/****************************************/

func (h *H1WarpHandler) NextWarpId() int {
	return FIXED_WARP_ID
}

func (h *H1WarpHandler) NewWarpData(warpId int) *warpship.WarpData {
	return warpship.NewWarpData(h.Ship(), warpId)
}

func (h *H1WarpHandler) VerifyProtocol(protocol string) exception2.IOException {
	return nil
}

/****************************************/
/* Implements H1CommandHandler         */
/****************************************/

func (h *H1WarpHandler) HandleHeader(cmd *CmdHeader) (common.NextSocketAction, exception2.IOException) {
	var ioerr exception2.IOException = nil

	for { // try catch
		wsip := h.Ship().(*impl.WarpShipImpl)
		var tur tour.Tour
		tur, ioerr = wsip.GetTour(FIXED_WARP_ID, true)
		if ioerr != nil {
			break
		}
		wdat := warpship.WarpDataGet(tur)

		baylog.Debug("%s handleHeader status=%d", wdat, cmd.status)
		wsip.Keeping = false
		if h.state == STATE_FINISHED {
			h.changeState(STATE_READ_HEADER)
		}
		if h.state != STATE_READ_HEADER {
			ioerr = exception.NewProtocolException("Header command not expected")
			break
		}

		if bayserver.Harbor().TraceHeader() {
			baylog.Info("%s warp_http: resStatus: %d", wdat, cmd.status)
		}

		for _, nv := range cmd.headers {
			tur.Res().Headers().Add(nv[0], nv[1])
			if bayserver.Harbor().TraceHeader() {
				baylog.Info("%s warp_http: resHeader: %s=%s", wdat, nv[0], nv[1])
			}
		}

		tur.Res().Headers().SetStatus(cmd.status)
		resContLen := tur.Res().Headers().ContentLength()
		ioerr = tur.Res().SendHeaders(tourimpl.TOUR_ID_NOCHECK)
		if ioerr != nil {
			break
		}

		if resContLen == 0 || cmd.status == httpstatus.NOT_MODIFIED {
			ioerr = h.endResContent(tur)
			if ioerr != nil {
				break
			}
		} else {
			h.changeState(STATE_READ_CONTENT)
			sid := wsip.ShipId()
			tur.Res().SetConsumeListener(func(length int, resume bool) {
				if resume {
					wsip.ResumeRead(sid)
				}
			})
		}

		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}

	return -1, ioerr
}

func (h *H1WarpHandler) HandleContent(cmd *CmdContent) (common.NextSocketAction, exception2.IOException) {
	var ioerr exception2.IOException = nil

	for { // try catch
		var tur tour.Tour
		tur, ioerr = h.Ship().GetTour(FIXED_WARP_ID, true)
		if ioerr != nil {
			break
		}

		wdat := warpship.WarpDataGet(tur)

		baylog.Debug("%s handleContent len=%d posted=%d contLen=%d", wdat, cmd.length, tur.Res().BytesPosted(), tur.Res().BytesLimit())

		if h.state != STATE_READ_CONTENT {
			ioerr = exception.NewProtocolException("Content command not expected")
		}

		available := true
		available, ioerr = tur.Res().SendResContent(tourimpl.TOUR_ID_NOCHECK, cmd.buffer, cmd.start, cmd.length)
		if ioerr != nil {
			break
		}

		if tur.Res().BytesPosted() == tur.Res().BytesLimit() {
			ioerr = h.endResContent(tur)
			if ioerr != nil {
				break
			}
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil

		} else if !available {
			return common.NEXT_SOCKET_ACTION_SUSPEND, nil

		} else {
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}
	}

	return -1, ioerr
}

func (h *H1WarpHandler) HandleEndContent(cmd *CmdEndContent) (common.NextSocketAction, exception2.IOException) {
	bayserver.FatalError(exception2.NewSink(""))
	return -1, nil
}

func (h *H1WarpHandler) ReqFinished() bool {
	return h.state == STATE_FINISHED
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (h *H1WarpHandler) endResContent(tur tour.Tour) exception2.IOException {
	h.Ship().EndWarpTour(tur, true)
	ioerr := tur.Res().EndResContent(tourimpl.TOUR_ID_NOCHECK)
	if ioerr != nil {
		return ioerr
	}
	h.resetState()
	h.Ship().(*impl.WarpShipImpl).Keeping = true
	return nil
}

func (h *H1WarpHandler) changeState(newState int) {
	h.state = newState
}

func (h *H1WarpHandler) resetState() {
	h.changeState(STATE_FINISHED)
}

func (h *H1WarpHandler) Ship() warpship.WarpShip {
	return h.protocolHandler.Ship().(warpship.WarpShip)
}
