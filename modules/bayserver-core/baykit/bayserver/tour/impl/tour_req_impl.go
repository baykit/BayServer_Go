package impl

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"encoding/base64"
	"regexp"
	"strconv"
	"strings"
)

type TourReqImpl struct {
	tour *TourImpl
	key  int

	uri      string
	method   string
	protocol string

	headers *headers.Headers

	rewrittenURI string
	queryString  string
	pathInfo     string
	scriptName   string
	reqHost      string
	reqPort      int

	remoteUser string
	remotePass string

	remoteAddress  string
	remotePort     int
	remoteHostFunc tour.RemoteHostResolver

	serverAddress string
	serverPort    int
	serverName    string
	charset       string

	bytesPosted    int
	bytesConsumed  int
	bytesLimit     int
	contentHandler tour.ReqContentHandler
	available      bool
	ended          bool
}

func NewTourReq(tur *TourImpl) *TourReqImpl {
	return &TourReqImpl{
		tour:    tur,
		headers: headers.NewHeaders(),
	}
}

func (req *TourReqImpl) Init(key int) {

}

func (req *TourReqImpl) String() string {
	return req.tour.String()
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (req *TourReqImpl) Reset() {
	//baylog.Info("TourReq:Reset")
	req.headers.Clear()
	req.key = 0
	req.uri = ""
	req.method = ""
	req.protocol = ""
	req.bytesPosted = 0
	req.bytesConsumed = 0
	req.bytesLimit = 0

	req.rewrittenURI = ""
	req.queryString = ""
	req.pathInfo = ""
	req.scriptName = ""
	req.reqHost = ""
	req.reqPort = 0
	req.remoteUser = ""
	req.remotePass = ""

	req.remoteAddress = ""
	req.remotePort = 0
	req.remoteHostFunc = nil
	req.serverAddress = ""
	req.serverPort = 0
	req.serverName = ""

	req.charset = ""
	req.contentHandler = nil
	req.available = false
	req.ended = false
}

/****************************************/
/* Implements TourReq                   */
/****************************************/

func (req *TourReqImpl) Key() int {
	return req.key
}

func (req *TourReqImpl) Uri() string {
	return req.uri
}

func (req *TourReqImpl) SetUri(uri string) {
	req.uri = uri
}

func (req *TourReqImpl) Method() string {
	return req.method
}

func (req *TourReqImpl) SetMethod(method string) {
	req.method = method
}

func (req *TourReqImpl) Protocol() string {
	return req.protocol
}

func (req *TourReqImpl) SetProtocol(protocol string) {
	req.protocol = protocol
}

func (req *TourReqImpl) ReqHost() string {
	return req.reqHost
}

func (req *TourReqImpl) ReqPort() int {
	return req.reqPort
}

func (req *TourReqImpl) RemoteAddress() string {
	return req.remoteAddress
}

func (req *TourReqImpl) SetRemoteAddress(adr string) {
	req.remoteAddress = adr
}

func (req *TourReqImpl) RemotePort() int {
	return req.remotePort
}

func (req *TourReqImpl) SetRemotePort(port int) {
	req.remotePort = port
}

func (req *TourReqImpl) RemoteHost() string {
	return req.remoteHostFunc()
}

func (req *TourReqImpl) SetRemoteHostFunc(resolver tour.RemoteHostResolver) {
	req.remoteHostFunc = resolver
}

func (req *TourReqImpl) ServerAddress() string {
	return req.serverAddress
}

func (req *TourReqImpl) SetServerAddress(adr string) {
	req.serverAddress = adr
}

func (req *TourReqImpl) ServerPort() int {
	return req.serverPort
}

func (req *TourReqImpl) SetServerPort(port int) {
	req.serverPort = port
}

func (req *TourReqImpl) ServerName() string {
	return req.serverName
}

func (req *TourReqImpl) SetServerName(name string) {
	req.serverName = name
}

func ConstructTourReq(req *TourReqImpl) {
	req.headers = headers.NewHeaders()
}

func (req *TourReqImpl) Headers() *headers.Headers {
	return req.headers
}

func (req *TourReqImpl) SetLimit(limit int) {
	req.bytesLimit = limit
	req.bytesConsumed = 0
	req.bytesPosted = 0
	req.available = true
}

func (req *TourReqImpl) QueryString() string {
	return req.queryString
}

func (req *TourReqImpl) SetQueryString(queryString string) {
	req.queryString = queryString
}

func (req *TourReqImpl) ScriptName() string {
	return req.scriptName
}

func (req *TourReqImpl) SetScriptName(name string) {
	req.scriptName = name
}

func (req *TourReqImpl) PathInfo() string {
	return req.pathInfo
}

func (req *TourReqImpl) SetPathInfo(pathInfo string) {
	req.pathInfo = pathInfo
}

func (req *TourReqImpl) Charset() string {
	return req.charset
}

func (req *TourReqImpl) SetCharset(charset string) {
	req.charset = charset
}

func (req *TourReqImpl) RewrittenUri() string {
	return req.rewrittenURI
}

func (req *TourReqImpl) SetRewrittenUri(uri string) {
	req.rewrittenURI = uri
}

func (req *TourReqImpl) GetReqContentHandler() tour.ReqContentHandler {
	return req.contentHandler
}

func (req *TourReqImpl) SetReqContentHandler(handler tour.ReqContentHandler) {
	if handler == nil {
		bayserver.FatalError(exception.NewSink("Content handler is nil"))
	}
	if req.contentHandler != nil {
		bayserver.FatalError(exception.NewSink("Content handler is already set"))
	}

	req.contentHandler = handler
}

/**
 * Parse AUTHORIZATION headerq
 */

func (req *TourReqImpl) ParseAuthorization() {
	auth := req.headers.Get(headers.AUTHORIZATION)
	if auth != "" {
		ptn := regexp.MustCompile(`^Basic (.*)$`)
		mch := ptn.FindStringSubmatch(auth)
		if mch == nil {
			baylog.Debug("Not matched with basic authentication format: %s", auth)

		} else {
			bytes, err := base64.StdEncoding.DecodeString(mch[1])
			if err != nil {
				baylog.DebugE(exception.NewIOExceptionFromError(err), "Base64 Decode Error: %s", auth)

			} else {

				auth = string(bytes)

				ptn = regexp.MustCompile("(.*):(.*)")
				mch = ptn.FindStringSubmatch(auth)
				if mch == nil {
					baylog.Debug("Not matched with basic authentication format: %s", auth)

				} else {
					req.remoteUser = mch[1]
					req.remotePass = mch[2]
				}
			}
		}
	}
}

func (req *TourReqImpl) ParseHostPort(defaultPort int) {
	req.reqHost = ""

	hostPort := req.headers.Get(headers.X_FORWARDED_HOST)
	if hostPort != "" {
		req.headers.Remove(headers.X_FORWARDED_HOST)
		req.headers.Set(headers.HOST, hostPort)
	}

	hostPort = req.headers.Get(headers.HOST)
	if hostPort != "" {
		pos := strings.LastIndex(hostPort, ":")
		if pos == -1 {
			req.reqHost = hostPort
			req.reqPort = defaultPort

		} else {
			var err error
			req.reqHost = hostPort[0:pos]
			req.reqPort, err = strconv.Atoi(hostPort[pos+1:])
			if err != nil {
				baylog.ErrorE(exception.NewIOExceptionFromError(err), "")
			}
		}
	}
}

/**
 * This method passes a part of the POST request's content to the ReqContentHandler.
 * Additionally, it reduces the internal buffer space by the size of the data passed
 */

func (req *TourReqImpl) PostReqContent(checkId int, data []byte, start int, len int, lis tour.ContentConsumeListener) (bool, exception2.HttpException) {
	req.tour.CheckTourId(checkId)

	dataPassed := false

	// If has error, only read content. (Do not call content handler)
	if req.tour.error != nil {
		baylog.Debug("%s tour has error.", req.tour)

	} else if !req.tour.IsReading() {
		return false, exception2.NewHttpException(httpstatus.BAD_REQUEST, "%s tour is not reading.", req.tour)

	} else if req.contentHandler == nil {
		baylog.Warn("%s content read, but no content handler", req.tour)

	} else if req.bytesPosted+len > req.bytesLimit {
		return false, exception2.NewHttpException(httpstatus.BAD_REQUEST, baymessage.Get(symbol.HTP_READ_DATA_EXCEEDED, req.bytesPosted+len, req.bytesLimit))

	} else {
		ioerr := req.contentHandler.OnReadReqContent(req.tour, data, start, len, lis)
		if ioerr != nil {
			baylog.DebugE(ioerr, "")
			return false, exception2.NewHttpException(httpstatus.BAD_REQUEST, "%s Error on call onReadReqContent", req.tour)
		}
		dataPassed = true
	}

	req.bytesPosted += len
	baylog.Debug("%s read content: len=%d posted=%d limit=%d consumed=%d available=%t",
		req.tour, len, req.bytesPosted, req.bytesLimit, req.bytesConsumed, req.available)

	if !dataPassed {
		return true, nil
	}

	oldAvailable := req.available
	if !req.bufferAvailable() {
		req.available = false
	}

	if oldAvailable && !req.available {
		baylog.Debug("%s request unavailable (_ _).zZZ: posted=%d consumed=%d", req, req.bytesPosted, req.bytesConsumed)
	}

	return req.available, nil
}

/**
 * When calling this method, it is uncertain whether the response will be synchronous or asynchronous.
 * If it is synchronous, the tour will be disposed, and no further processing on the tour will be permitted.
 */

func (req *TourReqImpl) EndReqContent(checkId int) (exception.IOException, exception2.HttpException) {
	baylog.Debug("%s endReqContent", req)
	req.tour.CheckTourId(checkId)
	if req.ended {
		bayserver.FatalError(exception.NewSink("%s Request content is already ended", req.tour))
	}
	req.tour.ChangeState(TOUR_ID_NOCHECK, STATE_RUNNING)
	req.ended = true

	if req.bytesLimit >= 0 && req.bytesPosted != req.bytesLimit {
		bayserver.FatalError(exception.NewSink("nvalid request data length: %d/%d", req.bytesPosted, req.bytesLimit))
	}

	if req.contentHandler != nil {
		ioerr, hterr := req.contentHandler.OnEndReqContent(req.tour)
		if ioerr != nil || hterr != nil {
			return ioerr, hterr
		}
	}

	return nil, nil
}

func (req *TourReqImpl) RemoteUser() string {
	return req.remoteUser
}

func (req *TourReqImpl) RemotePass() string {
	return req.remotePass
}

func (req *TourReqImpl) BytesPosted() int {
	return req.bytesPosted
}

func (req *TourReqImpl) BytesLimit() int {
	return req.bytesLimit
}

/**
 * This method is called when the content of a POST request is consumed by the ReqContentHandler.
 * It then increases the internal buffer space by the amount consumed
 */

func (req *TourReqImpl) Consumed(checkId int, length int, lis tour.ContentConsumeListener) {
	req.tour.CheckTourId(checkId)

	req.bytesConsumed += length
	baylog.Debug("%s reqConsumed: len=%d posted=%d limit=%d consumed=%d available=%b",
		req.tour, length, req.bytesPosted, req.bytesLimit, req.bytesConsumed, req.available)

	var resume = false

	var oldAvailable = req.available
	if req.bufferAvailable() {
		req.available = true
	}
	if !oldAvailable && req.available {
		baylog.Debug("%s request available (^o^): posted=%d consumed=%d", req, req.bytesPosted, req.bytesConsumed)
		resume = true
	}

	lis(length, resume)
}

func (req *TourReqImpl) Abort() bool {
	baylog.Debug("%s req abort", req)
	if req.tour.IsPreparing() {
		req.tour.ChangeState(req.tour.tourId, STATE_ABORTED)
		return true

	} else if req.tour.IsRunning() {
		aborted := true

		if req.contentHandler != nil {
			aborted = req.contentHandler.OnAbortReq(req.tour)
		}

		return aborted

	} else {
		baylog.Debug("%s tour is not preparing or not running", req.tour)
		return false
	}
}

func (req *TourReqImpl) bufferAvailable() bool {
	return req.bytesPosted-req.bytesConsumed < bayserver.Harbor().TourBufferSize()
}
