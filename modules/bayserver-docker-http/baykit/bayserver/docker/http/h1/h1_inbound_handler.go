package h1

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	common2 "bayserver-core/baykit/bayserver/common/inboundship/impl"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	ship "bayserver-core/baykit/bayserver/ship/impl"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"bayserver-core/baykit/bayserver/util/urlencoder"
	"bayserver-docker-http/baykit/bayserver/docker/http"
	"net"
	"strings"
)

const COMMAND_STATE_READ_HEADER = 1
const COMMAND_STATE_READ_CONTENT = 2
const COMMAND_STATE_READ_FINISHED = 3

type H1InboundHandler struct {
	protocolHandler *H1ProtocolHandlerImpl
	headerRead      bool
	httpProtocol    string
	state           int
	curReqId        int
	curTour         tour.Tour
	curTourId       int
}

func NewH1InboundHandler() *H1InboundHandler {
	h := &H1InboundHandler{
		curReqId: 1,
	}
	h.resetState()

	var _ tour.TourHandler = h // interface check
	return h
}

func (h *H1InboundHandler) Init(handler *H1ProtocolHandlerImpl) {
	h.protocolHandler = handler
}

func (h *H1InboundHandler) String() string {
	return "H1InboundHandler"
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *H1InboundHandler) Reset() {
	h.resetState()
	h.headerRead = false
	h.httpProtocol = ""
	h.curReqId = 1
	h.curTour = nil
	h.curTourId = 0
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (h *H1InboundHandler) OnProtocolError(err exception.ProtocolException) (bool, exception2.IOException) {
	baylog.DebugE(err, "")
	return true, nil
}

/****************************************/
/* Implements TourHandler          */
/****************************************/

func (h *H1InboundHandler) SendHeaders(tur tour.Tour) exception2.IOException {
	resCon := ""

	// determine Connection header value
	if tur.Req().Headers().GetConnection() != headers.CONNECTION_TYPE_KEEP_ALIVE {
		// If client doesn't support "Keep-Alive", set "Close"
		resCon = "Close"

	} else {
		resCon = "Keep-Alive"
		// Client supports "Keep-Alive"
		if tur.Res().Headers().GetConnection() != headers.CONNECTION_TYPE_KEEP_ALIVE {
			clen := tur.Res().Headers().ContentLength()

			// If tour doesn't need "Keep-Alive"
			if clen == -1 {
				// If content-length not specified
				if tur.Res().Headers().ContentType() != "" &&
					strings.HasPrefix(tur.Res().Headers().ContentType(), "text/") {
					// If content is text, connection must be closed
					resCon = "Close"
				}
			}
		}
	}

	tur.Res().Headers().Set(headers.CONNECTION, resCon)

	if bayserver.Harbor().TraceHeader() {
		baylog.Info("%s resStatus:%d", tur, tur.Res().Headers().Status())
		for _, name := range tur.Res().Headers().HeaderNames() {
			for _, value := range tur.Res().Headers().HeaderValues(name) {
				baylog.Info("%s resHeader: %s=%s", name, value)
			}
		}
	}

	cmd := NewResHeader(tur.Res().Headers(), tur.Req().Protocol())
	return h.protocolHandler.Post(cmd, nil)
}

func (h *H1InboundHandler) SendContent(tour tour.Tour, bytes []byte, ofs int, length int, lis common.DataConsumeListener) exception2.IOException {
	cmd := NewCmdContent(bytes, ofs, length)
	return h.protocolHandler.Post(cmd, lis)
}

func (h *H1InboundHandler) SendEnd(tour tour.Tour, keepAlive bool, lis common.DataConsumeListener) exception2.IOException {
	sip := h.Ship()
	baylog.Debug("%s H1 sendEnd: tur=%s keep=%t", sip, tour, keepAlive)

	// Send end request command
	cmd := NewCmdEndContent()
	sid := sip.ShipId()

	ensureFunc := func() {
		//baylog.Debug("%s Call back from post end. keep=%t", sip, keepAlive)
		if keepAlive {
			sip.Keeping = true
			sip.ResumeRead(sid)

		} else {
			sip.PostClose(sid)
		}
	}

	ioerr := h.protocolHandler.Post(cmd, func() {
		//baylog.Debug("%s call back of end content command: tur=%s", sip, tour)
		ensureFunc()
		lis()
	})

	if ioerr != nil {
		baylog.Error("Error: %s", ioerr)
		ensureFunc()
		return ioerr
	}

	return nil
}

/****************************************/
/* Implements H1CommandHandler          */
/****************************************/

func (h *H1InboundHandler) HandleHeader(cmd *CmdHeader) (common.NextSocketAction, exception2.IOException) {
	sip := h.Ship()
	baylog.Debug("%s handleHeader: method=%s uri=%s proto=%s", sip, cmd.method, cmd.uri, cmd.version)

	if h.state == COMMAND_STATE_READ_FINISHED {
		h.changeState(COMMAND_STATE_READ_HEADER)
	}

	if h.state != COMMAND_STATE_READ_HEADER || h.curTour != nil {
		err := exception.NewProtocolException("Header command not expected: state=%d curTour=%s", h.state, h.curTour)
		h.resetState()
		return -1, err
	}

	// check HTTP2
	protocol := strings.ToUpper(cmd.version)
	if protocol == "HTTP/2.0" {
		port := sip.PortDocker().(http.HtpPortDocker)
		if port.SupportH2() {
			sip.PortDocker().ReturnProtocolHandler(sip.AgentId(), h.protocolHandler)
			protocolHandler := protocolhandlerstore.GetStore(http.H2_PROTO_NAME, true, sip.AgentId()).Rent().(H1ProtocolHandler)
			sip.SetProtocolHandler(protocolHandler)
			return -1, exception.NewUpgradeException()
		}
	}

	tur := sip.GetTour(h.curReqId, false, true)
	if tur == nil {
		baylog.Error(baymessage.Get(symbol.INT_NO_MORE_TOURS))
		tur = sip.GetTour(h.curReqId, true, true)
		ioerr := tur.Res().SendError(impl.TOUR_ID_NOCHECK, httpstatus.SERVICE_UNAVAILABLE, "No available tours", nil)
		if ioerr != nil {
			return 0, ioerr
		}
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}

	h.curTour = tur
	h.curTourId = tur.TourId()
	h.curReqId++

	common2.GetInboundShipImpl(sip).Keeping = false

	tur.Req().SetUri(urlencoder.EncodeTilde(cmd.uri))
	tur.Req().SetMethod(strings.ToUpper(cmd.method))
	tur.Req().SetProtocol(protocol)

	if !(protocol == "HTTP/1.1") || (protocol == "HTTP/1.0") || (protocol == "HTTP/0.9") {
		return -1, exception.NewProtocolException(baymessage.Get(symbol.HTP_UNSUPPORTED_PROTOCOL, protocol))
	}

	for _, nv := range cmd.headers {
		tur.Req().Headers().Add(nv[0], nv[1])
	}

	reqContLen := tur.Req().Headers().ContentLength()
	baylog.Debug("%s read header method=%s protocol=%s uri=%s contlen=%d",
		sip, tur.Req().Method(), tur.Req().Protocol(), tur.Req().Uri(), reqContLen)

	if bayserver.Harbor().TraceHeader() {
		for _, nv := range cmd.headers {
			baylog.Info("%s h1: reqHeader: %s=%s", tur, nv[0], nv[1])
		}
	}

	if reqContLen > 0 {
		tur.Req().SetLimit(reqContLen)
	}

	var ioerr exception2.IOException = nil
	var hterr exception.HttpException = nil

	for { // try/catch
		hterr = h.startTour(tur)
		if hterr != nil {
			break
		}

		if reqContLen <= 0 {
			ioerr, hterr = h.endReqContent(h.curTourId, tur)
			if hterr != nil {
				break
			}
			if ioerr != nil {
				break
			}

			return common.NEXT_SOCKET_ACTION_SUSPEND, nil

		} else {
			h.changeState(STATE_READ_CONTENT)
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}
	}

	if ioerr != nil {
		return 0, ioerr

	} else {
		// hterr != nil

		baylog.DebugE(hterr, "%s Http error occurred: %v", h, hterr)
		if reqContLen <= 0 {
			// not post data
			ioerr = tur.Res().SendHttpException(impl.TOUR_ID_NOCHECK, hterr)
			if ioerr != nil {
				return 0, ioerr
			}

			h.resetState()
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil

		} else {
			// Delay send
			h.changeState(STATE_READ_CONTENT)
			tur.SetHttpError(hterr)
			tur.Req().SetReqContentHandler(&tour.DevNullContentHandler{})
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}

	}

}

func (h *H1InboundHandler) HandleContent(cmd *CmdContent) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s handleContent: len=%d", h.Ship(), cmd.length)

	if h.state != STATE_READ_CONTENT {
		s := h.state
		h.resetState()
		return -1, exception.NewProtocolException("Content command not expected: state=%d", s)
	}

	tur := h.curTour
	tourId := h.curTourId

	var ioerr exception2.IOException = nil
	var hterr exception.HttpException = nil
	for { // try-catch
		sid := h.Ship().ShipId()
		var success bool
		success, hterr = tur.Req().PostReqContent(
			tourId,
			cmd.buffer,
			cmd.start,
			cmd.length,
			func(length int, resume bool) {
				if resume {
					tur.Ship().(*ship.ShipImpl).ResumeRead(sid)
				}
			})

		if hterr != nil {
			break
		}

		if tur.Req().BytesPosted() == tur.Req().BytesLimit() {
			if tur.Error() != nil {
				// Error has occurred on header completed
				baylog.Debug("%s Delay send error", tur)
				hterr = tur.Error()
				break

			} else {
				ioerr, hterr = h.endReqContent(tourId, tur)
				if ioerr != nil || hterr != nil {
					break
				}
				return common.NEXT_SOCKET_ACTION_CONTINUE, nil
			}
		}

		if !success {
			return common.NEXT_SOCKET_ACTION_SUSPEND, nil // end reading
		} else {
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}

	}

	// error
	for { // try catch
		if hterr != nil {
			tur.Req().Abort()
			ioerr = tur.Res().SendHttpException(tourId, hterr)
			if ioerr != nil {
				break
			}
			h.resetState()
			return common.NEXT_SOCKET_ACTION_WRITE, nil
		}
	}

	// ioerr != nil
	return -1, ioerr
}

func (h *H1InboundHandler) HandleEndContent(cmd *CmdEndContent) (common.NextSocketAction, exception2.IOException) {
	bayserver.FatalError(exception2.NewSink(""))
	return -1, nil
}

func (h *H1InboundHandler) ReqFinished() bool {
	return h.state == COMMAND_STATE_READ_FINISHED
}

/****************************************/
/* Private functions                    */
/****************************************/

func (h *H1InboundHandler) Ship() *common2.InboundShipImpl {
	return h.protocolHandler.Ship().(*common2.InboundShipImpl)
}

func (h *H1InboundHandler) changeState(newState int) {
	h.state = newState
}

func (h *H1InboundHandler) resetState() {
	h.headerRead = false
	h.changeState(COMMAND_STATE_READ_FINISHED)
	h.curTour = nil
}

func (h *H1InboundHandler) startTour(tur tour.Tour) exception.HttpException {
	secure := h.Ship().PortDocker().Secure()
	defaultPort := 80
	if secure {
		defaultPort = 443
	}
	req := tur.Req()
	req.ParseHostPort(defaultPort)
	req.ParseAuthorization()

	// Get remote address
	clientAdr := req.Headers().Get(headers.X_FORWARDED_FOR)
	if clientAdr != "" {
		req.SetRemoteAddress(clientAdr)
		req.SetRemotePort(-1)

	} else {
		local := common2.GetInboundShipImpl(h.Ship()).Conn.LocalAddr().(*net.TCPAddr)
		remote := common2.GetInboundShipImpl(h.Ship()).Conn.RemoteAddr().(*net.TCPAddr)
		req.SetRemotePort(remote.Port)
		req.SetRemoteAddress(remote.IP.String())
		req.SetServerAddress(local.IP.String())
	}
	req.SetRemoteHostFunc(tour.NewDefaultRemoteHostResolver(req.RemoteAddress()))

	req.SetServerPort(req.ReqPort())
	req.SetServerName(req.ReqHost())
	tur.SetSecure(secure)

	return tur.Go()
}

func (h *H1InboundHandler) endReqContent(checkTourId int, tur tour.Tour) (exception2.IOException, exception.HttpException) {
	ioerr, hterr := tur.Req().EndReqContent(checkTourId)
	if ioerr != nil || hterr != nil {
		return ioerr, hterr
	}
	h.resetState()
	return nil, nil
}
