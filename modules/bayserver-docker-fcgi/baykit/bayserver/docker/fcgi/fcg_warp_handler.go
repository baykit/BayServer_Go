package fcgi

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
	"bayserver-core/baykit/bayserver/util/cgiutil"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"unicode"
)

const STATE_READ_HEADER = 1
const STATE_READ_CONTENT = 2

type FcgWarpHandler struct {
	protocolHandler *FcgProtocolHandlerImpl
	curWarpId       int
	state           int
	lineBuf         []byte

	// for read header/contents
	pos  int
	last int
	data []byte
}

func NewFcgWarpHandler() *FcgWarpHandler {
	h := &FcgWarpHandler{}
	h.resetState()

	var _ tour.TourHandler = h     // cast check
	var _ warpship.WarpHandler = h // cast check
	var _ FcgHandler = h           // cast check
	var _ util.Reusable = h        // cast check
	return h
}

func (h *FcgWarpHandler) Init(handler *FcgProtocolHandlerImpl) {
	h.protocolHandler = handler
}

func (h *FcgWarpHandler) String() string {
	return "FcgWarpHandler"
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (h *FcgWarpHandler) Reset() {
	h.resetState()
	h.lineBuf = h.lineBuf[:0]
	h.pos = 0
	h.last = 0
	h.data = nil
	h.curWarpId++
}

/****************************************/
/* Implements TourHandler               */
/****************************************/

func (h *FcgWarpHandler) SendHeaders(tur tour.Tour) exception2.IOException {
	ioerr := h.sendBeginReq(tur)
	if ioerr != nil {
		return ioerr
	}
	return h.sendParams(tur)
}

func (h *FcgWarpHandler) SendContent(tur tour.Tour, buf []byte, start int, length int, lis common.DataConsumeListener) exception2.IOException {
	return h.sendStdIn(tur, buf, start, length, lis)
}

func (h *FcgWarpHandler) SendEnd(tur tour.Tour, keepAlive bool, lis common.DataConsumeListener) exception2.IOException {
	return h.sendStdIn(tur, make([]byte, 0), 0, 0, lis)
}

func (h *FcgWarpHandler) OnProtocolError(e exception.ProtocolException) (bool, exception2.IOException) {
	bayserver.FatalError(exception2.NewSink(""))
	return false, nil
}

/****************************************/
/* Implements WarpHandler               */
/****************************************/

func (h *FcgWarpHandler) NextWarpId() int {
	h.curWarpId++
	return h.curWarpId
}

func (h *FcgWarpHandler) NewWarpData(warpId int) *warpship.WarpData {
	return warpship.NewWarpData(h.Ship(), warpId)
}

func (h *FcgWarpHandler) VerifyProtocol(protocol string) exception2.IOException {
	return nil
}

/****************************************/
/* Implements FcgCommandHandler         */
/****************************************/

func (h *FcgWarpHandler) HandleBeginRequest(cmd *CmdBeginRequest) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid FCGI command: %d", cmd.Type())
}

func (h *FcgWarpHandler) HandleEndRequest(cmd *CmdEndRequest) (common.NextSocketAction, exception2.IOException) {
	tur, perr := h.Ship().GetTour(cmd.ReqId(), true)
	if perr != nil {
		return -1, perr
	}
	h.endReqContent(tur)
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *FcgWarpHandler) HandleParams(cmd *CmdParams) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid FCGI command: %d", cmd.Type())
}

func (h *FcgWarpHandler) HandleStdErr(cmd *CmdStdErr) (common.NextSocketAction, exception2.IOException) {
	msg := string(cmd.Data[cmd.Start : cmd.Start+cmd.Length])
	baylog.Error("%s sserver error: %s", h, msg)
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (h *FcgWarpHandler) HandleStdIn(cmd *CmdStdIn) (common.NextSocketAction, exception2.IOException) {
	return -1, exception.NewProtocolException("Invalid FCGI command: %d", cmd.Type())
}

func (h *FcgWarpHandler) HandleStdOut(cmd *CmdStdOut) (common.NextSocketAction, exception2.IOException) {
	tur, perr := h.Ship().GetTour(cmd.ReqId(), true)
	if perr != nil {
		return -1, perr
	}
	if tur == nil {
		bayserver.FatalError(exception2.NewSink("Tour not found"))
	}

	if cmd.Length == 0 {
		// stdout end
		h.resetState()
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}

	h.data = cmd.Data
	h.pos = cmd.Start
	h.last = cmd.Start + cmd.Length

	var ioerr exception2.IOException = nil
	for { // try catch
		if h.state == STATE_READ_HEADER {
			ioerr = h.readHeader(tur)
			if ioerr != nil {
				break
			}
		}

		if h.pos < h.last {
			if h.state == STATE_READ_CONTENT {
				var available bool
				available, ioerr = tur.Res().SendResContent(tourimpl.TOUR_ID_NOCHECK, h.data, h.pos, h.last-h.pos)
				if ioerr != nil {
					break
				}
				if !available {
					return common.NEXT_SOCKET_ACTION_SUSPEND, nil
				}
			}
		}

		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	}

	return -1, ioerr
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (h *FcgWarpHandler) endResHeader(tur tour.Tour) exception2.IOException {
	wdat := warpship.WarpDataGet(tur)
	wdat.ResHeaders.CopyTo(tur.Res().Headers())
	ioerr := tur.Res().SendHeaders(tourimpl.TOUR_ID_NOCHECK)
	if ioerr != nil {
		return ioerr
	}
	h.changeState(STATE_READ_CONTENT)
	return nil
}

func (h *FcgWarpHandler) endReqContent(tur tour.Tour) exception2.IOException {
	h.Ship().EndWarpTour(tur, true)
	ioerr := tur.Res().EndResContent(tourimpl.TOUR_ID_NOCHECK)
	if ioerr != nil {
		return ioerr
	}
	h.resetState()
	return nil
}

func (h *FcgWarpHandler) changeState(newState int) {
	h.state = newState
}

func (h *FcgWarpHandler) resetState() {
	h.changeState(STATE_READ_HEADER)
}

func (h *FcgWarpHandler) readHeader(tur tour.Tour) exception2.IOException {
	wdat := warpship.WarpDataGet(tur)

	var ioerr exception2.IOException = nil
	for { // try catch

		var headerFinished bool
		headerFinished, ioerr = h.parseHeader(wdat.ResHeaders)
		if ioerr != nil {
			break
		}
		if headerFinished {
			wdat.ResHeaders.CopyTo(tur.Res().Headers())

			// Check HTTP Status from headers
			status := wdat.ResHeaders.Get(headers.STATUS)
			if status != "" {
				parts := strings.Split(status, " ")
				if len(parts) == 0 {
					ioerr = exception.NewProtocolException("warp: Status header of server is invalid: %s", status)
					break
				}

				stCode, err := strconv.Atoi(parts[0])
				if err != nil {
					baylog.ErrorE(exception2.NewExceptionFromError(err), "")
					ioerr = exception.NewProtocolException("warp: Status header of server is invalid: %s", status)
					break
				}

				tur.Res().Headers().SetStatus(stCode)
				tur.Req().Headers().Remove(headers.STATUS)
			}
			sip := h.Ship()

			baylog.Debug("%s fcgi: read header status=%s contlen=%d", sip, status, wdat.ResHeaders.ContentLength())
			sid := sip.ShipId()
			tur.Res().SetConsumeListener(func(length int, resume bool) {
				if resume {
					sip.ResumeRead(sid)
				}
			})

			ioerr = tur.Res().SendHeaders(tourimpl.TOUR_ID_NOCHECK)
			if ioerr != nil {
				break
			}

			h.changeState(STATE_READ_CONTENT)
		}

		return nil
	}

	return ioerr
}

func (h *FcgWarpHandler) parseHeader(headers *headers.Headers) (bool, exception2.IOException) {

	var ioerr exception2.IOException = nil

	for {
		if h.pos == h.last {
			// no byte data
			break
		}

		c := h.data[h.pos]
		h.pos++

		if c == '\r' {
			continue

		} else if c == '\n' {
			line := string(h.lineBuf)
			if line == "" {
				return true, nil
			}
			colonPos := strings.Index(line, ":")

			if colonPos < 0 {
				ioerr = exception.NewProtocolException("fcgi: Header line of server is invalid: %s", line)
				break

			} else {
				name := ""
				value := ""
				for p := colonPos - 1; p >= 0; p-- {
					// trimming header name
					if !unicode.IsSpace(rune(line[p])) {
						name = line[:p+1]
						break
					}
				}
				for p := colonPos + 1; p < len(line); p++ {
					// trimming header value
					if !unicode.IsSpace(rune(line[p])) {
						value = line[p:]
						break
					}
				}
				if name == "" || value == "" {
					ioerr = exception.NewProtocolException("fcgi: Header line of server is invalid: %s", line)
					break
				}
				headers.Add(name, value)
				if bayserver.Harbor().TraceHeader() {
					baylog.Info("%s fcgi_warp: resHeader: %s=%s", h.Ship(), name, value)
				}
			}
			h.lineBuf = h.lineBuf[:0]
		} else {
			h.lineBuf = append(h.lineBuf, c)
		}

	}

	if ioerr != nil {
		return false, ioerr
	}

	return true, nil
}

func (h *FcgWarpHandler) sendStdIn(tur tour.Tour, data []byte, ofs int, length int, lis common.DataConsumeListener) exception2.IOException {
	cmd := NewCmdStdIn2(warpship.WarpDataGet(tur).WarpId, data, ofs, length)
	return h.Ship().Post(cmd, lis)
}

func (h *FcgWarpHandler) sendBeginReq(tur tour.Tour) exception2.IOException {
	cmd := NewCmdBeginRequest(warpship.WarpDataGet(tur).WarpId)
	cmd.Role = FCGI_RESPONDER
	cmd.KeepCon = true
	return h.Ship().Post(cmd, nil)
}

func (h *FcgWarpHandler) sendParams(tur tour.Tour) exception2.IOException {
	var ioerr exception2.IOException = nil
	for { // try catch
		scriptBase := h.Ship().Docker().(FcgWarpDocker).ScriptBase()
		if scriptBase == "" {
			scriptBase = tur.Town().(docker.Town).Location()
		}

		if scriptBase == "" {
			ioerr = exception2.NewIOException("%s scriptBase of fcgi docker or location of town is not specified.", tur.Town())
			break
		}

		docRoot := h.Ship().Docker().(FcgWarpDocker).DocRoot()
		if docRoot == "" {
			docRoot = tur.Town().(docker.Town).Location()
		}

		if docRoot == "" {
			ioerr = exception2.NewIOException("%s docRoot of fcgi docker or location of town is not specified.", tur.Town())
			break
		}

		warpId := warpship.WarpDataGet(tur).WarpId
		cmd := NewCmdParams(warpId)

		var scriptFname = ""
		cgiutil.GetEnv2(tur.Town().(docker.Town).Name(), docRoot, scriptBase, tur, func(name string, value string) {
			if name == cgiutil.SCRIPT_FILENAME {
				scriptFname = value
			} else {
				cmd.addParam(name, value)
			}
		})

		scriptFname = "proxy:fcgi://" + h.Ship().Docker().Host() + ":" + strconv.Itoa(h.Ship().Docker().Port()) + scriptFname
		cmd.addParam(cgiutil.SCRIPT_FILENAME, scriptFname)

		cmd.addParam(CONTEXT_PREFIX, "")
		cmd.addParam(UNIQUE_ID, uuid.New().String())

		if bayserver.Harbor().TraceHeader() {
			for _, kv := range cmd.Params {
				baylog.Info("%s fcgi_warp: env: %s=%s", h.Ship(), kv[0], kv[1])
			}
		}

		ioerr = h.Ship().Post(cmd, nil)
		if ioerr != nil {
			break
		}

		cmdParamsEnd := NewCmdParams(warpId)
		ioerr = h.Ship().Post(cmdParamsEnd, nil)
		if ioerr != nil {
			break
		}

		break
	}

	return ioerr
}

func (h *FcgWarpHandler) Ship() warpship.WarpShip {
	return h.protocolHandler.Ship().(warpship.WarpShip)
}
