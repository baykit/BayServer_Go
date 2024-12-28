package ajp

import (
	"bayserver-core/baykit/bayserver/agent/monitor"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	common2 "bayserver-core/baykit/bayserver/common/inboundship/impl"
	impl2 "bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
)

const COMMAND_STATE_READ_FORWARD_REQUEST = 1
const COMMAND_STATE_READ_DATA = 2

const AJP_DUMMY_KEY = 1

type AjpInboundHandler struct {
	protocolHandler *AjpProtocolHandlerImpl
	curTourId       int
	reqCommand      *CmdForwardRequest
	state           int
	keeping         bool
}

func NewAjpInboundHandler() *AjpInboundHandler {
	h := &AjpInboundHandler{}
	h.resetState()

	var _ tour.TourHandler = h // interface check
	var _ AjpHandler = h       // interface check
	return h
}

func (h *AjpInboundHandler) Init(handler *AjpProtocolHandlerImpl) {
	h.protocolHandler = handler
}

func (h *AjpInboundHandler) String() string {
	return "AjpInboundHandler"
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *AjpInboundHandler) Reset() {
	h.resetState()
	h.reqCommand = nil
	h.keeping = false
	h.curTourId = 0
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (h *AjpInboundHandler) OnProtocolError(err exception.ProtocolException) (bool, exception2.IOException) {
	baylog.DebugE(err, "")
	ibShip := h.Ship()
	tur := ibShip.GetErrorTour()
	tur.Res().SendError(impl.TOUR_ID_NOCHECK, httpstatus.BAD_REQUEST, err.Error(), err)
	return true, nil
}

/****************************************/
/* Implements TourHandler          */
/****************************************/

func (h *AjpInboundHandler) SendHeaders(tur tour.Tour) exception2.IOException {
	baylog.Debug("%s PH:sendHeaders: tur=%s", h.Ship(), tur)

	cmd := NewCmdSendHeaders()
	for _, name := range tur.Res().Headers().HeaderNames() {
		for _, value := range tur.Res().Headers().HeaderValues(name) {
			cmd.AddHeader(name, value)
		}
	}
	cmd.SetStatus(tur.Res().Headers().Status())
	return h.protocolHandler.Post(cmd, nil)
}

func (h *AjpInboundHandler) SendContent(tur tour.Tour, bytes []byte, ofs int, length int, lis common.DataConsumeListener) exception2.IOException {
	cmd := NewCmdSendBodyChunk(bytes, ofs, length)
	return h.protocolHandler.Post(cmd, lis)
}

func (h *AjpInboundHandler) SendEnd(tur tour.Tour, keepAlive bool, lis common.DataConsumeListener) exception2.IOException {
	sip := h.Ship()
	baylog.Debug("%s AJP sendEnd: tur=%s keep=%t", sip, tur, keepAlive)

	cmd := NewCmdEndResponse()
	cmd.Reuse = keepAlive

	sid := sip.ShipId()
	ensureFunc := func() {
		baylog.Debug("%s Call back from post end. keep=%t", sip, keepAlive)
		if !keepAlive {
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

func (h *AjpInboundHandler) HandleForwardRequest(cmd *CmdForwardRequest) (common.NextSocketAction, exception2.IOException) {
	sip := h.Ship()
	baylog.Debug("%s handleForwardRequest: method=%s uri=%s", sip, cmd.Method, cmd.ReqUri)

	if h.state != COMMAND_STATE_READ_FORWARD_REQUEST {
		return -1, exception.NewProtocolException("Ajpi: Invalid command: %d state=%d", cmd.Type(), h.state)
	}

	h.keeping = false
	h.reqCommand = cmd
	tur := sip.GetTour(AJP_DUMMY_KEY, false, true)

	if tur == nil {
		baylog.Error(baymessage.Get(symbol.INT_NO_MORE_TOURS))
		tur = sip.GetTour(AJP_DUMMY_KEY, true, true)
		ioerr := tur.Res().SendError(impl.TOUR_ID_NOCHECK, httpstatus.SERVICE_UNAVAILABLE, "No available tours", nil)
		if ioerr != nil {
			return 0, ioerr
		}
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}

	h.curTourId = tur.TourId()
	tur.Req().SetUri(cmd.ReqUri)
	tur.Req().SetProtocol(cmd.Protocol)
	tur.Req().SetMethod(cmd.Method)
	cmd.Headers.CopyTo(tur.Req().Headers())

	queryString := cmd.Attributes["?query_string"]
	if queryString != "" {
		tur.Req().SetUri(tur.Req().Uri() + "?" + queryString)
	}

	baylog.Debug("%s read header method=%s protocol=%s uri=%s contlen=%d",
		tur, tur.Req().Method(), tur.Req().Protocol(), tur.Req().Uri(), tur.Req().Headers().ContentLength())
	if bayserver.Harbor().TraceHeader() {
		for _, name := range cmd.Headers.HeaderNames() {
			for _, value := range cmd.Headers.HeaderValues(name) {
				baylog.Info("%s header: %s=%s", tur, name, value)
			}
		}
	}

	reqContLen := cmd.Headers.ContentLength()

	if reqContLen > 0 {
		tur.Req().SetLimit(reqContLen)
	}

	var hterr exception.HttpException = nil
	var ioerr exception2.IOException = nil
	for {
		// try catch
		hterr = h.startTour(tur)
		if hterr != nil {
			break
		}

		if reqContLen <= 0 {
			ioerr, hterr = h.endReqContent(tur)
			if ioerr != nil || hterr != nil {
				break
			}

		} else {
			h.changeState(COMMAND_STATE_READ_DATA)
		}

		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}

	if hterr != nil {
		if reqContLen <= 0 {
			tur.Req().Abort()
			ioerr = tur.Res().SendHttpException(impl.TOUR_ID_NOCHECK, hterr)
			if ioerr != nil {
				return -1, ioerr
			}
			h.resetState()
			return common.NEXT_SOCKET_ACTION_WRITE, nil

		} else {
			// Delay send
			h.changeState(COMMAND_STATE_READ_DATA)
			tur.SetHttpError(hterr)
			tur.Req().SetReqContentHandler(tour.NewDevNullContentHandler())
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}
	}

	if ioerr != nil {
		return -1, ioerr
	}
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *AjpInboundHandler) HandleData(cmd *CmdData) (common.NextSocketAction, exception2.IOException) {
	sip := h.Ship()
	baylog.Debug("%s handleData len=%d", sip, cmd.Length)

	if h.state != COMMAND_STATE_READ_DATA {
		return -1, exception.NewProtocolException("Invalid Ajp Command: %d", cmd.Type())
	}

	tur := sip.GetTour(AJP_DUMMY_KEY, false, true)

	var hterr exception.HttpException = nil

	for {
		// try catch

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

		if tur.Req().BytesPosted() == tur.Req().BytesLimit() {
			// request content completed
			if tur.Error() != nil {
				// Error has occurred on header completed
				baylog.Debug("%s Delay send error", tur)
				hterr = tur.Error()
				break

			} else {
				var ioerr exception2.IOException
				ioerr, hterr = h.endReqContent(tur)
				if ioerr != nil {
					return -1, ioerr
				}
				if hterr != nil {
					break
				}
				return common.NEXT_SOCKET_ACTION_CONTINUE, nil
			}

		} else {
			bch := NewCmdGetBodyChunk()
			bch.ReqLen = tur.Req().BytesLimit() - tur.Req().BytesPosted()
			if bch.ReqLen > AJP_MAX_DATA_LEN {
				bch.ReqLen = AJP_MAX_DATA_LEN
			}
			ioerr := h.protocolHandler.Post(bch, nil)
			if ioerr != nil {
				return -1, ioerr
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

func (h *AjpInboundHandler) HandleEndResponse(cmd *CmdEndResponse) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid Ajp Command: %d", cmd.Type())
}

func (h *AjpInboundHandler) HandleSendBodyChunk(cmd *CmdSendBodyChunk) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid Ajp Command: %d", cmd.Type())
}

func (h *AjpInboundHandler) HandleSendHeaders(cmd *CmdSendHeaders) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid Ajp Command: %d", cmd.Type())
}

func (h *AjpInboundHandler) HandleShutdown(cmd *CmdShutdown) (common.NextSocketAction, exception2.IOException) {
	ioerr := monitor.ShutdownAll()
	if ioerr != nil {
		return -1, ioerr
	}
	return common.NEXT_SOCKET_ACTION_CLOSE, nil
}

func (h *AjpInboundHandler) HandleGetBodyChunk(cmd *CmdGetBodyChunk) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid Ajp Command: %d", cmd.Type())
}

func (h *AjpInboundHandler) NeedData() bool {
	return h.state == COMMAND_STATE_READ_DATA
}

/****************************************/
/* Private functions                    */
/****************************************/

func (h *AjpInboundHandler) Ship() *common2.InboundShipImpl {
	return h.protocolHandler.Ship().(*common2.InboundShipImpl)
}

func (h *AjpInboundHandler) changeState(newState int) {
	h.state = newState
}

func (h *AjpInboundHandler) resetState() {
	h.changeState(COMMAND_STATE_READ_FORWARD_REQUEST)
}

func (h *AjpInboundHandler) endReqContent(tur tour.Tour) (exception2.IOException, exception.HttpException) {
	ioerr, hterr := tur.Req().EndReqContent(impl.TOUR_ID_NOCHECK)
	if ioerr != nil || hterr != nil {
		return ioerr, hterr
	}
	h.resetState()
	return nil, nil
}

func (h *AjpInboundHandler) startTour(tur tour.Tour) exception.HttpException {
	req := tur.Req()
	if h.reqCommand.IsSsl {
		req.ParseHostPort(443)
	} else {
		req.ParseHostPort(80)
	}
	req.ParseAuthorization()

	req.SetRemotePort(-1)
	req.SetRemoteAddress(h.reqCommand.RemoteAddr)
	req.SetRemoteHostFunc(func() string { return h.reqCommand.RemoteHost })

	tur.Req().SetServerAddress(impl2.GetConn(h.Ship().Rudder()).LocalAddr().String())
	req.SetServerPort(h.reqCommand.ServerPort)
	req.SetServerName(h.reqCommand.ServerName)
	tur.SetSecure(h.reqCommand.IsSsl)

	return tur.Go()
}
