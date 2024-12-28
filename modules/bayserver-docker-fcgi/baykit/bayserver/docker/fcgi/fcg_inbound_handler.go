package fcgi

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	common2 "bayserver-core/baykit/bayserver/common/inboundship/impl"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/cgiutil"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"bayserver-core/baykit/bayserver/util/httputil"
	"bytes"
	"strconv"
	"strings"
)

const COMMAND_STATE_READ_BEGIN_REQUEST = 1
const COMMAND_STATE_READ_PARAMS = 2
const COMMAND_STATE_READ_STDIN = 3

type FcgInboundHandler struct {
	protocolHandler *FcgProtocolHandlerImpl
	state           int
	env             map[string]string
	reqId           int
	reqKeepAlive    bool
}

func NewFcgInboundHandler() *FcgInboundHandler {
	h := &FcgInboundHandler{
		env: map[string]string{},
	}
	h.resetState()

	var _ tour.TourHandler = h // interface check
	return h
}

func (h *FcgInboundHandler) Init(handler *FcgProtocolHandlerImpl) {
	h.protocolHandler = handler
}

func (h *FcgInboundHandler) String() string {
	return "FcgInboundHandler"
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *FcgInboundHandler) Reset() {
	h.resetState()
	h.env = map[string]string{}
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (h *FcgInboundHandler) OnProtocolError(err exception.ProtocolException) (bool, exception2.IOException) {
	baylog.DebugE(err, "")
	ibShip := h.Ship()
	tur := ibShip.GetErrorTour()
	tur.Res().SendError(impl.TOUR_ID_NOCHECK, httpstatus.BAD_REQUEST, err.Error(), err)
	return true, nil
}

/****************************************/
/* Implements TourHandler          */
/****************************************/

func (h *FcgInboundHandler) SendHeaders(tur tour.Tour) exception2.IOException {
	baylog.Debug("%s PH:sendHeaders: tur=%s", h.Ship(), tur)

	scode := tur.Res().Headers().Status()
	status := strconv.Itoa(scode) + " " + httpstatus.GetDescription(scode)
	tur.Res().Headers().Set(headers.STATUS, status)

	if bayserver.Harbor().TraceHeader() {
		baylog.Info("%s resStatus:%d", tur, tur.Res().Headers().Status())
		for _, name := range tur.Res().Headers().HeaderNames() {
			for _, value := range tur.Res().Headers().HeaderValues(name) {
				baylog.Info("%s resHeader: %s=%s", name, value)
			}
		}
	}

	var buf bytes.Buffer
	ioerr := httputil.SendMimeHeaders(tur.Res().Headers(), &buf)
	if ioerr != nil {
		return ioerr
	}
	ioerr = httputil.SendNewLine(&buf)
	if ioerr != nil {
		return ioerr
	}

	cmd := NewCmdStdOut(tur.Req().Key(), buf.Bytes(), 0, buf.Len())
	return h.protocolHandler.Post(cmd, nil)
}

func (h *FcgInboundHandler) SendContent(tur tour.Tour, bytes []byte, ofs int, length int, lis common.DataConsumeListener) exception2.IOException {
	cmd := NewCmdStdOut(tur.Req().Key(), bytes, ofs, length)
	return h.protocolHandler.Post(cmd, lis)
}

func (h *FcgInboundHandler) SendEnd(tur tour.Tour, keepAlive bool, lis common.DataConsumeListener) exception2.IOException {
	sip := h.Ship()
	baylog.Debug("%s Fcg sendEnd: tur=%s keep=%t", sip, tur, keepAlive)

	// Send empty stdout command
	stdOutCmd := NewCmdStdOut(tur.Req().Key(), nil, 0, 0)
	ioerr := h.protocolHandler.Post(stdOutCmd, nil)
	if ioerr != nil {
		return ioerr
	}

	// Send end request command
	endCmd := NewCmdEndRequest(tur.Req().Key())
	sid := sip.ShipId()
	ensureFunc := func() {
		//baylog.Debug("%s Call back from post end. keep=%t", sip, keepAlive)
		if !keepAlive {
			sip.PostClose(sid)
		}
	}

	ioerr = h.protocolHandler.Post(endCmd, func() {
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

func (h *FcgInboundHandler) HandleBeginRequest(cmd *CmdBeginRequest) (common.NextSocketAction, exception2.IOException) {
	sip := h.Ship()
	baylog.Debug("%s handleBeginRequest: reqId=%d keep=%t", sip, cmd.ReqId(), cmd.KeepCon)

	if h.state != COMMAND_STATE_READ_BEGIN_REQUEST {
		return -1, exception.NewProtocolException("fcgi: Invalid command: %d state=%d", cmd.Type(), h.state)
	}

	reqId := cmd.ReqId()
	h.checkReqId(reqId)
	tur := sip.GetTour(reqId, false, true)

	if tur == nil {
		baylog.Error(baymessage.Get(symbol.INT_NO_MORE_TOURS))
		tur = sip.GetTour(reqId, true, true)
		ioerr := tur.Res().SendError(impl.TOUR_ID_NOCHECK, httpstatus.SERVICE_UNAVAILABLE, "No available tours", nil)
		if ioerr != nil {
			return 0, ioerr
		}
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}

	h.reqKeepAlive = cmd.KeepCon

	h.changeState(COMMAND_STATE_READ_PARAMS)
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *FcgInboundHandler) HandleEndRequest(cmd *CmdEndRequest) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid FCGI Command: %d", cmd.Type())
}

func (h *FcgInboundHandler) HandleParams(cmd *CmdParams) (common.NextSocketAction, exception2.IOException) {
	sip := h.Ship()
	baylog.Debug("%s handleParams reqId=%d nParams=%d", sip, cmd.ReqId(), len(cmd.Params))

	if h.state != COMMAND_STATE_READ_PARAMS {
		return -1, exception.NewProtocolException("Invalid FCGI Command: %d", cmd.Type())
	}

	h.checkReqId(cmd.ReqId())

	tur := sip.GetTour(cmd.ReqId(), false, true)

	var ioerr exception2.IOException = nil
	for { // try-catch

		if len(cmd.Params) == 0 {
			// Header completed

			// check keep-alive
			//  keep-alive flag of BeginRequest has high priority
			if h.reqKeepAlive {
				if !tur.Req().Headers().Contains(headers.CONNECTION) {
					tur.Req().Headers().Set(headers.CONNECTION, "Keep-Alive")
				}

			} else {
				tur.Req().Headers().Set(headers.CONNECTION, "Close")
			}

			contLen := tur.Req().Headers().ContentLength()

			// end params
			baylog.Debug("%s handleBeginRequest: method=%s protocol=%s uri=%s contlen=%s",
				sip, tur.Req().Method(), tur.Req().Protocol(), tur.Req().Uri(), contLen)

			if bayserver.Harbor().TraceHeader() {
				for _, name := range tur.Req().Headers().HeaderNames() {
					for _, value := range tur.Req().Headers().HeaderValues(name) {
						baylog.Info("%s reqHeader: %s=%s", name, value)
					}
				}
			}

			if contLen > 0 {
				tur.Req().SetLimit(contLen)
			}

			h.changeState(COMMAND_STATE_READ_STDIN)

			hterr := h.startTour(tur)

			if hterr != nil {
				baylog.Debug("%s Http error occurred: %s", sip, hterr)
				if contLen <= 0 {
					tur.Req().Abort()
					// no post data
					ioerr = tur.Res().SendHttpException(impl.TOUR_ID_NOCHECK, hterr)
					if ioerr != nil {
						break
					}
					return common.NEXT_SOCKET_ACTION_CONTINUE, nil

				} else {
					// Delay send
					h.changeState(COMMAND_STATE_READ_STDIN)
					tur.SetHttpError(hterr)
					tur.Req().SetReqContentHandler(tour.NewDevNullContentHandler())
					return common.NEXT_SOCKET_ACTION_CONTINUE, nil
				}

			}
		} else {

			for _, nv := range cmd.Params {
				name := nv[0]
				value := nv[1]
				h.env[name] = value

				if strings.HasPrefix(name, "HTTP_") {
					hname := name[5:]
					tur.Req().Headers().Add(hname, value)

				} else if name == "CONTENT_TYPE" {
					tur.Req().Headers().Add(headers.CONTENT_TYPE, value)

				} else if name == "CONTENT_LENGTH" {
					tur.Req().Headers().Add(headers.CONTENT_LENGTH, value)

				} else if name == "HTTPS" {
					tur.SetSecure(strings.ToLower(value) == "on")

				}
			}

			tur.Req().SetUri(h.env["REQUEST_URI"])
			tur.Req().SetProtocol(h.env["SERVER_PROTOCOL"])
			tur.Req().SetMethod(h.env["REQUEST_METHOD"])

			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}

		break
	}
	// ioerr != nil
	return -1, ioerr
}

func (h *FcgInboundHandler) HandleStdErr(cmd *CmdStdErr) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid FCGI Command: %d", cmd.Type())
}

func (h *FcgInboundHandler) HandleStdIn(cmd *CmdStdIn) (common.NextSocketAction, exception2.IOException) {
	sip := h.Ship()
	baylog.Debug("%s handleStdIn reqId=%d len=%d", sip, cmd.ReqId(), cmd.Length)

	if h.state != COMMAND_STATE_READ_STDIN {
		return -1, exception.NewProtocolException("Invalid FCGI Command: %d", cmd.Type())
	}

	tur := sip.GetTour(cmd.ReqId(), false, true)

	var hterr exception.HttpException = nil

	for {
		// try catch
		h.checkReqId(cmd.ReqId())

		if cmd.Length == 0 {
			// request content completed
			if tur.Error() != nil {
				// Error has occurred on header completed
				baylog.Debug("%s Delay send error", tur)
				hterr = tur.Error()
				break

			} else {
				var ioerr exception2.IOException
				ioerr, hterr = h.endReqContent(impl.TOUR_ID_NOCHECK, tur)
				if ioerr != nil {
					return -1, ioerr
				}
				if hterr != nil {
					break
				}
			}

		} else {
			sid := sip.ShipId()
			var success bool
			success, hterr = tur.Req().PostReqContent(
				impl.TOUR_ID_NOCHECK,
				cmd.Data,
				cmd.Start,
				cmd.Length,
				func(len int, resume bool) {
					if resume {
						sip.ResumeRead(sid)
					}
				})

			if hterr != nil {
				break
			}

			if !success {
				return common.NEXT_SOCKET_ACTION_SUSPEND, nil
			} else {
				return common.NEXT_SOCKET_ACTION_CONTINUE, nil
			}

		}

		break
	}

	if hterr != nil {
		tur.Req().Abort()
		tur.Res().SendHttpException(impl.TOUR_ID_NOCHECK, hterr)
		h.resetState()
		return common.NEXT_SOCKET_ACTION_WRITE, nil
	}

	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *FcgInboundHandler) HandleStdOut(cmd *CmdStdOut) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid FCGI Command: %d", cmd.Type())
}

/****************************************/
/* Private functions                    */
/****************************************/

func (h *FcgInboundHandler) Ship() *common2.InboundShipImpl {
	return h.protocolHandler.Ship().(*common2.InboundShipImpl)
}

func (h *FcgInboundHandler) changeState(newState int) {
	h.state = newState
}

func (h *FcgInboundHandler) resetState() {
	h.changeState(COMMAND_STATE_READ_BEGIN_REQUEST)
	h.reqId = FCGI_NULL_REQUEST_ID
}

func (h *FcgInboundHandler) checkReqId(receivedId int) exception2.IOException {
	if receivedId == FCGI_NULL_REQUEST_ID {
		return exception.NewProtocolException("Invalid request id: %d", receivedId)
	}

	if h.reqId == FCGI_NULL_REQUEST_ID {
		h.reqId = receivedId
	}

	if h.reqId != receivedId {
		baylog.Error("%s invalid request id: received=%d reqId=%d", h.Ship(), receivedId, h.reqId)
		return exception.NewProtocolException("Invalid request id: %d", receivedId)
	}

	return nil
}

func (h *FcgInboundHandler) endReqContent(checkTourId int, tur tour.Tour) (exception2.IOException, exception.HttpException) {
	ioerr, hterr := tur.Req().EndReqContent(checkTourId)
	if ioerr != nil || hterr != nil {
		return ioerr, hterr
	}
	h.resetState()
	return nil, nil
}

func (h *FcgInboundHandler) startTour(tur tour.Tour) exception.HttpException {
	secure := h.Ship().PortDocker().Secure()
	defaultPort := 80
	if secure {
		defaultPort = 443
	}
	req := tur.Req()
	req.ParseHostPort(defaultPort)
	req.ParseAuthorization()

	remotePort, err := strconv.Atoi(h.env[cgiutil.REMOTE_PORT])
	if err != nil {
		baylog.ErrorE(err, "")
	}
	req.SetRemotePort(remotePort)

	req.SetRemoteAddress(h.env[cgiutil.REMOTE_ADDR])
	req.SetRemoteHostFunc(tour.NewDefaultRemoteHostResolver(req.RemoteAddress()))

	req.SetServerName(req.ReqHost())
	req.SetServerPort(req.ReqPort())

	var serverPort int
	serverPort, err = strconv.Atoi(h.env[cgiutil.SERVER_PORT])
	if err != nil {
		baylog.ErrorE(err, "")
		serverPort = 80
	}
	req.SetServerPort(serverPort)

	return tur.Go()
}
