package ajp

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/warpship"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/tour"
	tourimpl "bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"strings"
)

const FIXED_WARP_ID = 1

const STATE_READ_HEADER = 1
const STATE_READ_CONTENT = 2

type AjpWarpHandler struct {
	protocolHandler *AjpProtocolHandlerImpl
	state           int
	contReadLen     int
}

func NewAjpWarpHandler() *AjpWarpHandler {
	h := &AjpWarpHandler{}
	h.resetState()

	var _ tour.TourHandler = h     // cast check
	var _ warpship.WarpHandler = h // cast check
	var _ AjpHandler = h           // cast check
	var _ util.Reusable = h        // cast check
	return h
}

func (h *AjpWarpHandler) Init(handler *AjpProtocolHandlerImpl) {
	h.protocolHandler = handler
}

func (h *AjpWarpHandler) String() string {
	return h.Ship().String()
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *AjpWarpHandler) Reset() {
	h.resetState()
	h.contReadLen = 0
}

/****************************************/
/* Implements TourHandler               */
/****************************************/

func (h *AjpWarpHandler) SendHeaders(tur tour.Tour) exception2.IOException {
	return h.sendForwardRequest(tur)
}

func (h *AjpWarpHandler) SendContent(tur tour.Tour, buf []byte, start int, length int, lis common.DataConsumeListener) exception2.IOException {
	return h.sendData(tur, buf, start, length, lis)
}

func (h *AjpWarpHandler) SendEnd(tour tour.Tour, keepAlive bool, lis common.DataConsumeListener) exception2.IOException {
	return h.Ship().Post(nil, lis)
}

func (h *AjpWarpHandler) OnProtocolError(e exception.ProtocolException) (bool, exception2.IOException) {
	bayserver.FatalError(exception2.NewSink(""))
	return false, nil
}

/****************************************/
/* Implements WarpHandler               */
/****************************************/

func (h *AjpWarpHandler) NextWarpId() int {
	return 1
}

func (h *AjpWarpHandler) NewWarpData(warpId int) *warpship.WarpData {
	return warpship.NewWarpData(h.Ship(), warpId)
}

func (h *AjpWarpHandler) VerifyProtocol(protocol string) exception2.IOException {
	return nil
}

/****************************************/
/* Implements AjpCommandHandler         */
/****************************************/

func (h *AjpWarpHandler) HandleData(cmd *CmdData) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid AJP command: %d", cmd.Type())
}

func (h *AjpWarpHandler) HandleEndResponse(cmd *CmdEndResponse) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s handleEndResponse reuse=%t", h, cmd.Reuse)
	wsip := h.Ship()

	var ioerr exception2.IOException = nil
	for { // try-catch
		tur, perr := wsip.GetTour(FIXED_WARP_ID, true)
		if perr != nil {
			ioerr = perr
			break
		}

		if h.state == STATE_READ_HEADER {
			ioerr = h.endResHeader(tur)
			if ioerr != nil {
				break
			}
		}

		ioerr = h.endResContent(tur, cmd.Reuse)
		if ioerr != nil {
			break
		}

		if cmd.Reuse {
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		} else {
			return common.NEXT_SOCKET_ACTION_CLOSE, nil
		}

	}

	return -1, ioerr
}

func (h *AjpWarpHandler) HandleForwardRequest(cmd *CmdForwardRequest) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid AJP command: %d", cmd.Type())
}

func (h *AjpWarpHandler) HandleSendBodyChunk(cmd *CmdSendBodyChunk) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s handleSendBodyChunk len=%d", h, cmd.Length)
	wsip := h.Ship()

	var ioerr exception2.IOException = nil
	for { // try-catch
		tur, perr := wsip.GetTour(FIXED_WARP_ID, true)
		if perr != nil {
			ioerr = perr
			break
		}

		if h.state == STATE_READ_HEADER {

			sid := wsip.ShipId()
			tur.Res().SetConsumeListener(func(len int, resume bool) {
				if resume {
					wsip.ResumeRead(sid)
				}
			})

			ioerr = h.endResHeader(tur)
			if ioerr != nil {
				break
			}
		}

		var available bool
		available, ioerr = tur.Res().SendResContent(tur.TourId(), cmd.Chunk, 0, cmd.Length)
		if ioerr != nil {
			break
		}
		h.contReadLen += cmd.Length

		if available {
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		} else {
			return common.NEXT_SOCKET_ACTION_SUSPEND, nil
		}

	}

	return -1, ioerr
}

func (h *AjpWarpHandler) HandleSendHeaders(cmd *CmdSendHeaders) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s handleSendHeaders", h)

	var ioerr exception2.IOException = nil
	for { // try-catch
		tur, perr := h.Ship().GetTour(FIXED_WARP_ID, true)
		if perr != nil {
			ioerr = perr
			break
		}

		if h.state != STATE_READ_HEADER {
			ioerr = exception.NewProtocolException("Invalid AJP command: %d state=%s", cmd.Type(), h.state)
			break
		}

		wdata := warpship.WarpDataGet(tur)

		if bayserver.Harbor().TraceHeader() {
			baylog.Info("%s recv res status: %d", wdata, cmd.Status)
		}

		wdata.ResHeaders.SetStatus(cmd.Status)
		for name, values := range cmd.Headers {
			for _, value := range values {
				if bayserver.Harbor().TraceHeader() {
					baylog.Info("%s recv res header: %s=%s", wdata, name, value)
				}
				wdata.ResHeaders.Add(name, value)
			}
		}

		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}

	return -1, ioerr
}

func (h *AjpWarpHandler) HandleShutdown(cmd *CmdShutdown) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid AJP command: %d", cmd.Type())
}

func (h *AjpWarpHandler) HandleGetBodyChunk(cmd *CmdGetBodyChunk) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s handleGetBodyChunk", h)
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *AjpWarpHandler) NeedData() bool {
	return false
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (h *AjpWarpHandler) endResHeader(tur tour.Tour) exception2.IOException {
	wdat := warpship.WarpDataGet(tur)
	wdat.ResHeaders.CopyTo(tur.Res().Headers())
	ioerr := tur.Res().SendHeaders(tourimpl.TOUR_ID_NOCHECK)
	if ioerr != nil {
		return ioerr
	}
	h.changeState(STATE_READ_CONTENT)
	return nil
}

func (h *AjpWarpHandler) endResContent(tur tour.Tour, keep bool) exception2.IOException {
	h.Ship().EndWarpTour(tur, keep)
	ioerr := tur.Res().EndResContent(tourimpl.TOUR_ID_NOCHECK)
	if ioerr != nil {
		return ioerr
	}
	h.resetState()
	return nil
}

func (h *AjpWarpHandler) changeState(newState int) {
	h.state = newState
}

func (h *AjpWarpHandler) resetState() {
	h.changeState(STATE_READ_HEADER)
}

func (h *AjpWarpHandler) sendForwardRequest(tur tour.Tour) exception2.IOException {
	baylog.Debug("%s %s construct header", h, tur)
	wsip := h.Ship()

	cmd := NewCmdForwardRequest()
	cmd.SetToServer(true)
	cmd.Method = tur.Req().Method()
	cmd.Protocol = tur.Req().Protocol()
	relUri := ""
	if tur.Req().RewrittenUri() != "" {
		relUri = tur.Req().RewrittenUri()
	} else {
		relUri = tur.Req().Uri()
	}
	twnPath := tur.Town().(docker.Town).Name()
	if !strings.HasSuffix(twnPath, "/") {
		twnPath += "/"
	}
	relUri = relUri[len(twnPath):]
	reqUri := wsip.Docker().DestTown() + relUri

	pos := strings.Index(reqUri, "?")
	if pos >= 0 {
		cmd.ReqUri = reqUri[0:pos]
		cmd.Attributes["?query_string"] = reqUri[pos+1:]

	} else {
		cmd.ReqUri = reqUri
	}
	cmd.RemoteAddr = tur.Req().RemoteAddress()
	cmd.RemoteHost = tur.Req().RemoteHost()
	cmd.ServerName = tur.Req().ServerName()
	cmd.ServerPort = tur.Req().ServerPort()
	cmd.IsSsl = tur.Secure()
	tur.Req().Headers().CopyTo(cmd.Headers)
	cmd.ServerPort = wsip.Docker().Port()

	if bayserver.Harbor().TraceHeader() {
		for _, name := range cmd.Headers.HeaderNames() {
			for _, value := range cmd.Headers.HeaderValues(name) {
				baylog.Info("%s sendWarpHeader: %s=%s", warpship.WarpDataGet(tur), name, value)
			}
		}
	}
	ioerr := h.Ship().Post(cmd, nil)
	return ioerr
}

func (h *AjpWarpHandler) sendData(tur tour.Tour, data []byte, ofs int, length int, lis common.DataConsumeListener) exception2.IOException {
	baylog.Debug("%s construct contents", tur)
	wsip := h.Ship()

	cmd := NewCmdData(data, ofs, length)
	cmd.SetToServer(true)
	ioerr := wsip.Post(cmd, lis)
	return ioerr
}

func (h *AjpWarpHandler) Ship() warpship.WarpShip {
	return h.protocolHandler.Ship().(warpship.WarpShip)
}
