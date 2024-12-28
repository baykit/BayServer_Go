package file

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/agent/multiplexer"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/rudder"
	impl2 "bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"bayserver-core/baykit/bayserver/util/mimes"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"os"
	"strings"
)

type FileContentHandler struct {
	tour.ReqContentHandler
	path      string
	abortable bool
}

func NewFileContentHandler(path string) *FileContentHandler {
	return &FileContentHandler{
		path: path,
	}
}

/****************************************/
/* Implements ReqContentHandler         */
/****************************************/

func (h *FileContentHandler) OnReadReqContent(
	tour tour.Tour,
	buf []byte,
	start int,
	length int,
	lis tour.ContentConsumeListener) exception.IOException {

	baylog.Debug("%s onReadContent(Ignore) len=%d", tour, length)
	tour.Req().Consumed(tour.TourId(), length, lis)
	return nil
}

func (h *FileContentHandler) OnEndReqContent(tour tour.Tour) (exception.IOException, exception2.HttpException) {
	baylog.Debug("%s endContent", tour)
	hterr := h.sendFileAsync(tour, h.path, tour.Res().Charset())
	if hterr != nil {
		return nil, hterr
	}
	h.abortable = false
	return nil, nil
}

func (h *FileContentHandler) OnAbortReq(tour tour.Tour) bool {
	baylog.Debug("%s onAbort aborted=%s", tour, h.abortable)
	return h.abortable
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (h *FileContentHandler) sendFileAsync(tur tour.Tour, file string, charset string) exception2.HttpException {

	if sysutil.IsDirectory(file) {
		return exception2.NewHttpException(httpstatus.FORBIDDEN, file)
	} else if !sysutil.IsFile(file) {
		return exception2.NewHttpException(httpstatus.NOT_FOUND, file)
	}

	mimeType := ""

	rname := file
	pos := strings.LastIndex(rname, ".")
	if pos >= 0 {
		ext := strings.ToLower(rname[pos+1:])
		mimeType = mimes.Get(ext)
	}

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	if strings.HasPrefix(mimeType, "text/") && charset != "" {
		mimeType += "; charset=" + charset
	}

	var ioerr exception.IOException = nil
	var f *os.File = nil
	for { // try catch
		var err error
		f, err = os.Open(file)
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(ioerr)
			break
		}

		var fileSize int64
		fileSize, ioerr = sysutil.GetFileSize(f)
		if ioerr != nil {
			break
		}
		tur.Res().Headers().SetContentType(mimeType)
		tur.Res().Headers().SetContentLength(int(fileSize))
		ioerr = tur.Res().SendHeaders(impl.TOUR_ID_NOCHECK)
		if ioerr != nil {
			break
		}

		agt := agent.Get(tur.Ship().(ship.Ship).AgentId())
		var rd rudder.Rudder
		var mpx common.Multiplexer = nil

		mpxType := bayserver.Harbor().FileMultiplexer()
		switch mpxType {
		case docker.MULTI_PLEXER_TYPE_JOB:
			rd = impl2.NewFileRudder(f)
			mpx = agt.JobMultiplexer()

		default:
			bayserver.FatalError(exception.NewSink("Multiplexer type not supported: %s", docker.GetMultiplexerTypeName(mpxType)))
		}

		sip := NewSendFileShip()
		tp := multiplexer.NewPlainTransporter(mpx, sip, false, 8192, false)

		sip.Init(rd, tp, tur, int(fileSize))
		sid := sip.ShipId()
		tur.Res().SetConsumeListener(func(len int, resume bool) {
			if resume {
				sip.ResumeRead(sid)
			}
		})

		agt.NetMultiplexer().AddRudderState(rd, common.NewRudderState(rd, tp))
		agt.NetMultiplexer().ReqRead(rd)

		break
	}

	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
		return exception2.NewHttpException(httpstatus.INTERNAL_SERVER_ERROR, file)
	}

	return nil
}
