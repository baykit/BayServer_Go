package cgi

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/rudder"
	ship2 "bayserver-core/baykit/bayserver/ship/impl"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"strconv"
	"strings"
)

type CgiStdOutShip struct {
	ship2.ShipImpl
	fileWroteLen int
	tour         tour.Tour
	tourId       int
	handler      *CgiReqContentHandler

	remain        string
	headerReading bool
}

func NewCgiStdOutShip() *CgiStdOutShip {
	res := CgiStdOutShip{}
	res.ShipImpl.Construct()
	return &res
}

func (sip *CgiStdOutShip) Init(rd rudder.Rudder, agentId int, tur tour.Tour, tp common.Transporter, handler *CgiReqContentHandler) {
	sip.ShipImpl.Init(agentId, rd, tp)
	sip.handler = handler
	sip.tour = tur
	sip.tourId = tur.TourId()
	sip.headerReading = true
}

func (sip *CgiStdOutShip) String() string {
	return "agt#" + strconv.Itoa(sip.AgentId()) + " out_ship#" + strconv.Itoa(sip.ShipId()) + "/" + strconv.Itoa(sip.ObjectId())
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (sip *CgiStdOutShip) Reset() {
	sip.ShipImpl.Reset()
	sip.fileWroteLen = 0
	sip.tourId = 0
	sip.tour = nil
	sip.headerReading = true
	sip.remain = ""
	sip.handler = nil
}

/****************************************/
/* Implements Ship                      */
/****************************************/

func (sip *CgiStdOutShip) NotifyHandshakeDone(protocol string) (common.NextSocketAction, exception.IOException) {
	bayserver.FatalError(exception.NewSink(""))
	return 0, nil
}

func (sip *CgiStdOutShip) NotifyConnect() (common.NextSocketAction, exception.IOException) {
	bayserver.FatalError(exception.NewSink(""))
	return 0, nil
}

func (sip *CgiStdOutShip) NotifyRead(buf []byte) (common.NextSocketAction, exception.IOException) {
	sip.fileWroteLen += len(buf)
	baylog.Debug("%s read %d bytes: total=%d", sip, len(buf), sip.fileWroteLen)

	var ioerr exception.IOException = nil
	if sip.headerReading {
		for {
			pos := -1
			for i := 0; i < len(buf); i++ {
				// ByteArray.get(int) method does not increment position
				if buf[i] == '\n' {
					pos = i
					break
				}
			}
			//BayLog.debug("%s pos: %d", this, pos);

			if pos == -1 {
				break
			}

			line := string(buf[:pos])
			if len(sip.remain) > 0 {
				line = sip.remain + line
				sip.remain = ""
			}
			buf = buf[pos+1:]

			line = strings.TrimSpace(line)
			//BayLog.debug("line: %s", line);

			//  if line is empty ("\r\n")
			//  finish header reading.
			if line == "" {
				sip.headerReading = false
				ioerr = sip.tour.Res().SendHeaders(sip.tourId)
				break
			} else {
				if bayserver.Harbor().TraceHeader() {
					baylog.Info("%s CGI: res header line: %s", sip.tour, line)
				}

				sepPos := strings.Index(line, ":")
				if sepPos >= 0 {
					key := strings.TrimSpace(line[0:sepPos])
					val := strings.TrimSpace(line[sepPos+1:])
					if strings.ToLower(key) == "status" {
						tokens := strings.Fields(val)
						if len(tokens) < 1 {
							baylog.ErrorE(exception.NewException("Invalid status line: ", val), "")

						} else {
							n, err := strconv.Atoi(tokens[0])
							if err != nil {
								baylog.ErrorE(exception.NewExceptionFromError(err), "")
							} else {
								sip.tour.Res().Headers().SetStatus(n)
							}
						}

					} else {
						sip.tour.Res().Headers().Add(key, val)
					}
				}
			}
		}
	}

	if ioerr != nil {
		return 0, ioerr
	}

	available := true

	if sip.headerReading {
		sip.remain += string(buf)

	} else {
		if len(buf) > 0 {
			available, ioerr = sip.tour.Res().SendResContent(sip.tourId, buf, 0, len(buf))
			if ioerr != nil {
				sip.NotifyError(ioerr)
				return common.NEXT_SOCKET_ACTION_CLOSE, nil
			}
		}
	}

	sip.handler.Access()
	if available {
		return common.NEXT_SOCKET_ACTION_CONTINUE, nil
	} else {
		return common.NEXT_SOCKET_ACTION_SUSPEND, nil
	}

}

func (sip *CgiStdOutShip) NotifyEof() common.NextSocketAction {
	baylog.Debug("%s EOF", sip)
	return common.NEXT_SOCKET_ACTION_CLOSE
}

func (sip *CgiStdOutShip) NotifyError(e exception.Exception) {
	baylog.DebugE(e, "%s CGI notifyError", sip)
}

func (sip *CgiStdOutShip) NotifyProtocolError(e exception2.ProtocolException) (bool, exception.IOException) {
	bayserver.FatalError(exception.NewSink(""))
	return false, nil
}

func (sip *CgiStdOutShip) NotifyClose() {
	sip.handler.StdOutClosed()
}

func (sip *CgiStdOutShip) CheckTimeout(durationSec int) bool {
	if sip.handler.Timeout() {
		// Kill cgi cmd instead of handing timeout
		baylog.Warn("%s Kill cmd!: %s", sip.tour, sip.handler.cmd)
		sip.handler.cmd.Process.Kill()
		return true
	}
	return false
}
