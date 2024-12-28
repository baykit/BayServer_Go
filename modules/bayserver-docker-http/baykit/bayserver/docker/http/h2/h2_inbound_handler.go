package h2

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	common2 "bayserver-core/baykit/bayserver/common/inboundship/impl"
	impl2 "bayserver-core/baykit/bayserver/rudder/impl"
	ship2 "bayserver-core/baykit/bayserver/ship"
	ship "bayserver-core/baykit/bayserver/ship/impl"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/tour/tourstore"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"bayserver-docker-http/baykit/bayserver/docker/http/h2/h2_error_code"
	"net"
	"strconv"
	"strings"
)

const COMMAND_STATE_READ_HEADER = 1
const COMMAND_STATE_READ_CONTENT = 2
const COMMAND_STATE_READ_FINISHED = 3

type H2InboundHandler struct {
	protocolHandler *H2ProtocolHandlerImpl
	headerRead      bool
	httpProtocol    string

	reqContLen   int
	reqContRead  int
	windowSize   int
	settings     *H2Settings
	analyzer     *HeaderBlockAnalyzer
	reqHeaderTbl *HeaderTable
	resHeaderTbl *HeaderTable
}

func NewH2InboundHandler() *H2InboundHandler {
	h := &H2InboundHandler{}
	h.windowSize = bayserver.Harbor().TourBufferSize()
	h.settings = NewH2Settings()
	h.analyzer = NewHeaderBlockAnalyzer()
	h.reqHeaderTbl = CreateDynamicTable()
	h.resHeaderTbl = CreateDynamicTable()

	var _ tour.TourHandler = h // implement check
	var _ H2Handler = h        // implement check
	var _ H2CommandHandler = h // implement check
	return h
}

func (h *H2InboundHandler) Init(handler *H2ProtocolHandlerImpl) {
	h.protocolHandler = handler
}

func (h *H2InboundHandler) String() string {
	return "H2InboundHandler"
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *H2InboundHandler) Reset() {
	h.headerRead = false
	h.httpProtocol = ""
	h.reqContLen = 0
	h.reqContRead = 0
}

/****************************************/
/* Implements ProtocolHandler           */
/****************************************/

func (h *H2InboundHandler) OnProtocolError(err exception.ProtocolException) (bool, exception2.IOException) {
	baylog.DebugE(err, "")
	baylog.ErrorE(err, err.Error())
	cmd := NewCmdGoAway(CTL_STREAM_ID, nil)
	cmd.streamId = 0
	cmd.LastStreamId = 0
	cmd.ErrorCode = h2_error_code.PROTOCOL_ERROR
	cmd.DebugData = []byte("Thank you!")

	ioerr := h.protocolHandler.Post(cmd, nil)
	if ioerr != nil {
		return false, ioerr
	}
	h.protocolHandler.Ship().PostClose(ship2.SHIP_ID_NOCHECK)
	return true, nil
}

/****************************************/
/* Implements TourHandler          */
/****************************************/

func (h *H2InboundHandler) SendHeaders(tur tour.Tour) exception2.IOException {
	cmd := NewCmdHeaders(tur.Req().Key(), nil)

	bld := NewHeaderBlockBuilder()

	blk, perr := bld.BuildHeaderBlock(":status", strconv.Itoa(tur.Res().Headers().Status()), h.resHeaderTbl)
	if perr != nil {
		return perr
	}
	cmd.HeaderBlocks = append(cmd.HeaderBlocks, blk)

	// headers
	if bayserver.Harbor().TraceHeader() {
		baylog.Info("%s H2 res status: %d", tur, tur.Res().Headers().Status())
	}
	for _, name := range tur.Res().Headers().HeaderNames() {
		if strings.ToLower(name) == "connection" {
			baylog.Trace("%s Connection header is discarded", tur)

		} else {
			for _, value := range tur.Res().Headers().HeaderValues(name) {
				if bayserver.Harbor().TraceHeader() {
					baylog.Info("%s H2 res header: %s=%s", tur, name, value)
				}
				blk, perr = bld.BuildHeaderBlock(name, value, h.resHeaderTbl)
				if perr != nil {
					return perr
				}
				cmd.HeaderBlocks = append(cmd.HeaderBlocks, blk)
			}
		}
	}

	cmd.flags.SetEndHeaders(true)
	cmd.Excluded = false
	// cmd.streamDependency = streamId;
	cmd.flags.SetPadded(false)

	return h.protocolHandler.Post(cmd, nil)
}

func (h *H2InboundHandler) SendContent(tur tour.Tour, bytes []byte, ofs int, length int, lis common.DataConsumeListener) exception2.IOException {
	cmd := NewCmdData(tur.Req().Key(), nil, bytes, ofs, length)
	return h.protocolHandler.Post(cmd, lis)
}

func (h *H2InboundHandler) SendEnd(tur tour.Tour, keepAlive bool, lis common.DataConsumeListener) exception2.IOException {
	cmd := NewCmdData(tur.Req().Key(), nil, []byte{0}, 0, 0)
	cmd.flags.SetEndStream(true)
	return h.protocolHandler.Post(cmd, lis)
}

/****************************************/
/* Implements H1CommandHandler          */
/****************************************/

func (h *H2InboundHandler) HandlePreface(cmd *CmdPreface) (common.NextSocketAction, exception2.IOException) {
	sip := h.Ship()
	baylog.Debug("%s h2: handle_preface: proto=%s", sip, cmd.Protocol)

	h.httpProtocol = cmd.Protocol

	set := NewCmdSettings(CTL_STREAM_ID, nil)
	set.streamId = 0
	set.items = append(set.items, NewCmdSettingItem(MAX_CONCURRENT_STREAMS, tourstore.MAX_TOURS))
	set.items = append(set.items, NewCmdSettingItem(INITIAL_WINDOW_SIZE, h.windowSize))
	ioerr := h.protocolHandler.Post(set, nil)
	if ioerr != nil {
		return -1, ioerr
	}

	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *H2InboundHandler) HandleHeaders(cmd *CmdHeaders) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s handle_headers: stm=%d dep=%d weight=%d", h.Ship(), cmd.streamId, cmd.StreamDependency, cmd.Weight)
	var ioerr exception2.IOException = nil

catch:
	for { // try catch
		t := h.getTour(cmd.streamId)
		if t == nil {
			baylog.Error(baymessage.Get(symbol.INT_NO_MORE_TOURS))
			t = h.Ship().GetTour(cmd.streamId, true, true)
			ioerr = t.(tour.Tour).Res().SendError(impl.TOUR_ID_NOCHECK, httpstatus.SERVICE_UNAVAILABLE, "No available tours", nil)
			if ioerr != nil {
				break
			}
			//sip.agent.shutdown(false);
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}

		tur := t.(tour.Tour)
		for _, blk := range cmd.HeaderBlocks {
			if blk.op == HEADER_OP_UPDATE_DYNAMIC_TABLE_SIZE {
				baylog.Trace("%s header block update table size: %d", t, blk.size)
				h.reqHeaderTbl.SetSize(blk.size)
				continue
			}

			ioerr = h.analyzer.AnalyzeHeaderBlock(blk, h.reqHeaderTbl)
			if ioerr != nil {
				break catch
			}
			if bayserver.Harbor().TraceHeader() {
				baylog.Info("%s req header: %s=%s :%s", t, h.analyzer.Name, h.analyzer.Value, blk)
			}

			if h.analyzer.Name == "" {
				continue

			} else if h.analyzer.Name[0] != ':' {
				tur.Req().Headers().Add(h.analyzer.Name, h.analyzer.Value)

			} else if h.analyzer.Method != "" {
				tur.Req().SetMethod(h.analyzer.Method)

			} else if h.analyzer.Path != "" {
				tur.Req().SetUri(h.analyzer.Path)

			} else if h.analyzer.Scheme != "" {

			} else if h.analyzer.Status != "" {
				ioerr = exception2.NewIOException("Illegal state")
				break catch
			}
		}

		if cmd.flags.IsEndHeaders() {
			tur.Req().SetProtocol("HTTP/2.0")
			baylog.Debug("%s H2 read header method=%s protocol=%s uri=%s contlen=%d",
				h.Ship(), tur.Req().Method(), tur.Req().Protocol(), tur.Req().Uri(), tur.Req().Headers().ContentLength())

			reqContLen := tur.Req().Headers().ContentLength()

			if reqContLen > 0 {
				tur.Req().SetLimit(reqContLen)
			}

			var hterr exception.HttpException = nil
			for { // try-catch
				hterr = h.startTour(t)
				if hterr != nil {
					break
				}
				if tur.Req().Headers().ContentLength() <= 0 {
					ioerr, hterr = h.endReqContent(tur.TourId(), tur)
					if ioerr != nil {
						break catch
					}
					if hterr != nil {
						break
					}
				}
				break
			}

			if hterr != nil {
				baylog.Debug("%s Http error occurred: %s", h, hterr)
				if reqContLen <= 0 {
					// no post data
					tur.Req().Abort()
					ioerr = tur.Res().SendHttpException(impl.TOUR_ID_NOCHECK, hterr)
					if ioerr != nil {
						break catch
					}

					return common.NEXT_SOCKET_ACTION_CONTINUE, nil

				} else {
					// Delay send
					tur.SetHttpError(hterr)
					tur.Req().SetReqContentHandler(tour.NewDevNullContentHandler())
					return common.NEXT_SOCKET_ACTION_CONTINUE, nil
				}
			}
		}
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}
	return -1, ioerr
}

func (h *H2InboundHandler) HandleData(cmd *CmdData) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s handle_data: stm=%d len=%d", h.Ship(), cmd.streamId, cmd.length)
	tur := h.getTour(cmd.streamId).(*impl.TourImpl)
	if tur == nil {
		return -1, exception2.NewIOException("Invalid stream id: %d", cmd.streamId)
	}
	if tur.Req().Headers().ContentLength() <= 0 {
		return -1, exception.NewProtocolException("Post content not allowed")
	}

	var ioerr exception2.IOException = nil
	var hterr exception.HttpException = nil

	for { // try-catch
		success := true
		if cmd.length > 0 {
			tid := tur.TourId()
			success, hterr = tur.Req().PostReqContent(
				impl.TOUR_ID_NOCHECK,
				cmd.data,
				cmd.start,
				cmd.length,
				func(length int, resume bool) {
					tur.CheckTourId(tid)

					baylog.Debug("%s Callback from PostReqContent len=%d", h, length)
					if length > 0 {
						upd := NewCmdWindowUpdate(cmd.streamId, nil)
						upd.WindowSizeIncrement = length

						upd2 := NewCmdWindowUpdate(0, nil)
						upd2.WindowSizeIncrement = length

						var ioerror exception2.IOException = nil
						for { //try-catch
							ioerror = h.protocolHandler.Post(upd, nil)
							if ioerror != nil {
								break
							}
							ioerror = h.protocolHandler.Post(upd2, nil)
							if ioerror != nil {
								break
							}

							if resume {
								tur.Ship().(*ship.ShipImpl).ResumeRead(tur.ShipId())
							}
							return
						}

						baylog.ErrorE(ioerror, "")
					}

				})

			if hterr != nil {
				break
			}

			baylog.Debug("posted=%d contlen=%d", tur.Req().BytesPosted(), tur.Req().Headers().ContentLength())
			if tur.Req().BytesPosted() >= tur.Req().Headers().ContentLength() {

				if tur.Error() != nil {
					// Error has occurred on header completed
					baylog.Debug("%s Delay send error", tur)
					hterr = tur.Error()
					break

				} else {
					ioerr, hterr = h.endReqContent(tur.TourId(), tur)
					if ioerr != nil || hterr != nil {
						break
					}
				}
			}
		}

		if !success {
			return common.NEXT_SOCKET_ACTION_SUSPEND, nil
		} else {
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}

	}

	if hterr != nil {
		tur.Req().Abort()
		ioerr = tur.Res().SendHttpException(impl.TOUR_ID_NOCHECK, hterr)
		if ioerr != nil {
			return common.NEXT_SOCKET_ACTION_CONTINUE, nil
		}
	}

	return -1, ioerr
}

func (h *H2InboundHandler) HandlePriority(cmd *CmdPriority) (common.NextSocketAction, exception2.IOException) {
	if cmd.streamId == 0 {
		return -1, exception.NewProtocolException("Invalid streamId")
	}
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *H2InboundHandler) HandleSettings(cmd *CmdSettings) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s handleSettings: stmid=%d", h.Ship(), cmd.streamId)
	if cmd.flags.IsAck() {
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil // ignore ACK

	}

	for _, item := range cmd.items {
		baylog.Debug("%s handle: Setting id=%d, value=%d", h.Ship(), item.id, item.value)
		switch item.id {
		case HEADER_TABLE_SIZE:
			h.settings.HeaderTableSize = item.value

		case ENABLE_PUSH:
			h.settings.EnablePush = item.value != 0

		case MAX_CONCURRENT_STREAMS:
			h.settings.MaxConcurrentStreams = item.value

		case INITIAL_WINDOW_SIZE:
			h.settings.InitialWindowSize = item.value

		case MAX_FRAME_SIZE:
			h.settings.MaxFrameSize = item.value

		case MAX_HEADER_LIST_SIZE:
			h.settings.MaxHeaderListSize = item.value

		default:
			baylog.Debug("Invalid settings id (Ignore): %d", item.id)
		}
	}

	res := NewCmdSettings(0, NewH2Flags(FLAGS_ACK))
	ioerr := h.protocolHandler.Post(res, nil)
	if ioerr != nil {
		return -1, ioerr
	}

	return common.NEXT_SOCKET_ACTION_CONTINUE, nil

}

func (h *H2InboundHandler) HandleWindowUpdate(cmd *CmdWindowUpdate) (common.NextSocketAction, exception2.IOException) {
	if cmd.WindowSizeIncrement == 0 {
		return -1, exception.NewProtocolException("Invalid increment value")
	}
	baylog.Debug("%s handleWindowUpdate: stmid=%d siz=%d", h.Ship(), cmd.streamId, cmd.WindowSizeIncrement)
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *H2InboundHandler) HandleGoAway(cmd *CmdGoAway) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s received GoAway: lastStm=%d code=%d desc=%s debug=%s",
		h.Ship(), cmd.LastStreamId, cmd.ErrorCode, h2_error_code.Msg.Get(strconv.Itoa(cmd.ErrorCode)), string(cmd.DebugData))
	return common.NEXT_SOCKET_ACTION_CLOSE, nil
}

func (h *H2InboundHandler) HandlePing(cmd *CmdPing) (common.NextSocketAction, exception2.IOException) {
	sip := h.Ship()
	baylog.Debug("%s handle_ping: stm=%d", sip, cmd.streamId)

	res := NewCmdPing(cmd.streamId, NewH2Flags(FLAGS_ACK), cmd.opaqueData)
	ioerr := h.protocolHandler.Post(res, nil)
	if ioerr != nil {
		return -1, ioerr
	}
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *H2InboundHandler) HandleRstStream(cmd *CmdRstStream) (common.NextSocketAction, exception2.IOException) {
	baylog.Debug("%s received RstStream: stmid=%d code=%d desc=%s",
		h.Ship(), cmd.streamId, cmd.ErrorCode, h2_error_code.Msg.Get(strconv.Itoa(cmd.ErrorCode)))
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

/****************************************/
/* Private functions                    */
/****************************************/

func (h *H2InboundHandler) Ship() *common2.InboundShipImpl {
	return h.protocolHandler.Ship().(*common2.InboundShipImpl)
}

func (h *H2InboundHandler) getTour(key int) tour.Tour {
	return h.Ship().GetTour(key, false, true)
}

func (h *H2InboundHandler) startTour(tur tour.Tour) exception.HttpException {
	secure := h.Ship().PortDocker().Secure()
	defaultPort := 80
	if secure {
		defaultPort = 443
	}
	req := tur.Req()
	req.ParseHostPort(defaultPort)
	req.ParseAuthorization()

	tur.Req().SetProtocol(h.httpProtocol)

	// Get remote address
	clientAdr := req.Headers().Get(headers.X_FORWARDED_FOR)
	if clientAdr != "" {
		req.SetRemoteAddress(clientAdr)
		req.SetRemotePort(-1)

	} else {
		nrd := h.Ship().Rudder().(*impl2.TcpConnRudder)
		local := nrd.Conn.LocalAddr().(*net.TCPAddr)
		remote := nrd.Conn.RemoteAddr().(*net.TCPAddr)
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

func (h *H2InboundHandler) endReqContent(checkTourId int, tur tour.Tour) (exception2.IOException, exception.HttpException) {
	return tur.Req().EndReqContent(checkTourId)
}
