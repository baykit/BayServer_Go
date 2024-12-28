package impl

import (
	"bayserver-core/baykit/bayserver/bayserver"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/charutil"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	bytes2 "bytes"
	"fmt"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	"io"
)

type TourResImpl struct {
	tour       *TourImpl
	headers    *headers.Headers
	charset    string
	headerSent bool

	available          bool
	bytesPosted        int
	bytesConsumed      int
	bytesLimit         int
	resConsumeListener tour.ContentConsumeListener
	canCompress        bool
}

func NewTourRes(tur *TourImpl) *TourResImpl {
	r := &TourResImpl{
		tour:    tur,
		headers: headers.NewHeaders(),
	}
	var _ tour.TourRes = r // interface check
	return r
}

func (res *TourResImpl) String() string {
	return res.tour.String()
}

func (res *TourResImpl) Init() {

}

func (res *TourResImpl) Reset() {
	res.headers.Clear()
	res.bytesPosted = 0
	res.bytesConsumed = 0
	res.bytesLimit = 0

	res.charset = ""
	res.headerSent = false
	res.available = false
	res.resConsumeListener = nil
	res.canCompress = false
}

/****************************************/
/* Implements TourRes                   */
/****************************************/

func (res *TourResImpl) Charset() string {
	return res.charset
}

func (res *TourResImpl) SetCharset(charset string) {
	res.charset = charset
}

func (res *TourResImpl) Headers() *headers.Headers {
	return res.headers
}

func (res *TourResImpl) HeaderSent() bool {
	return res.headerSent
}

/**
 * This method sends the response headers to the client.
 * Whether this process is carried out synchronously or asynchronously is uncertain"
 */

func (res *TourResImpl) SendHeaders(checkId int) exception.IOException {
	tour := res.tour
	tour.CheckTourId(checkId)
	baylog.Debug("%s send header", res)

	if tour.IsZombie() || tour.IsAborted() {
		return nil
	}

	if res.headerSent {
		return nil
	}

	res.bytesLimit = res.headers.ContentLength()
	baylog.Debug("%s content length: %d", res, res.bytesLimit)

	var ioerr exception.IOException = nil

catch:
	for { // try-catch
		handled := false
		if !tour.errorHandling && res.headers.Status() >= 400 {
			trb := bayserver.Harbor().Trouble()
			if trb != nil {
				cmd := trb.Find(res.headers.Status())
				if cmd != nil {
					errTour := tour.ship.GetErrorTour().(*TourImpl)
					errTour.req.uri = cmd.Target
					tour.req.headers.CopyTo(errTour.req.headers)
					res.headers.CopyTo(errTour.res.headers)
					errTour.req.remotePort = tour.req.remotePort
					errTour.req.remoteAddress = tour.req.remoteAddress
					errTour.req.serverAddress = tour.req.serverAddress
					errTour.req.serverPort = tour.req.serverPort
					errTour.req.serverName = tour.req.serverName
					errTour.res.headerSent = tour.res.headerSent
					errTour.res.SetConsumeListener(func(len int, resume bool) {})
					tour.ChangeState(TOUR_ID_NOCHECK, STATE_ZOMBIE)
					switch cmd.Method {
					case docker.TROUBLE_METHOD_GUIDE:
						{
							hterr := errTour.Go()
							if hterr != nil {
								baylog.ErrorE(hterr, "")
								ioerr = exception.NewIOException(hterr.Error())
								break catch
							}
							break
						}

					case docker.TROUBLE_METHOD_TEXT:
						{
							data := []byte(cmd.Target)
							errTour.res.headers.SetContentLength(len(data))
							ioerr = errTour.res.SendHeaders(TOUR_ID_NOCHECK)
							if ioerr != nil {
								break catch
							}

							_, ioerr = errTour.res.SendResContent(TOUR_ID_NOCHECK, data, 0, len(data))
							if ioerr != nil {
								break catch
							}

							ioerr = errTour.res.EndResContent(TOUR_ID_NOCHECK)
							if ioerr != nil {
								break catch
							}
							break
						}

					case docker.TROUBLE_METHOD_REROUTE:
						{
							ioerr = errTour.res.SendHttpException(TOUR_ID_NOCHECK, exception2.NewMovedTemporarily(cmd.Target))
							if ioerr != nil {
								break catch
							}
							break
						}
					}
					handled = true
				}
			}
		}

		if !handled {
			ioerr = tour.ship.SendHeaders(tour.shipId, tour)
			if ioerr != nil {
				break catch
			}
		}

		break
	}

	res.headerSent = true

	if ioerr != nil {
		tour.ChangeState(checkId, STATE_ABORTED)
		return ioerr
	}

	return nil
}

/**
 * This method sends a part of the response content to the client.
 * Whether this process is synchronous or asynchronous is uncertain
 */

func (res *TourResImpl) SendResContent(checkId int, buf []byte, ofs int, length int) (bool, exception.IOException) {
	res.tour.CheckTourId(checkId)
	baylog.Debug("%s send content: len=%d", res, length)

	lis := func() {
		res.consumed(checkId, length)
	}

	if res.tour.IsZombie() {
		baylog.Debug("%s zombie return", res)
		lis()
		return true, nil
	}

	if !res.headerSent {
		bayserver.FatalError(exception.NewSink("Header not sent"))
		return false, nil
	}

	res.bytesPosted += length
	baylog.Debug("%s posted res content len=%d posted=%d limit=%d consumed=%d",
		res.tour, length, res.bytesPosted, res.bytesLimit, res.bytesConsumed)

	if res.tour.IsAborted() {
		// Don't send peer any data. Do nothing
		baylog.Debug("%s Aborted or zombie tour. do nothing: %s state=%s", res, res.tour, res.tour.state)
		res.tour.ChangeState(checkId, STATE_ENDED)
		lis()

	} else {
		if res.canCompress {

		} else {
			ioerr := res.tour.ship.SendResContent(res.tour.shipId, res.tour, buf, ofs, length, lis)

			if ioerr != nil {
				lis()
				res.tour.ChangeState(TOUR_ID_NOCHECK, STATE_ABORTED)
				return false, ioerr
			}
		}
	}

	if res.bytesLimit > 0 && res.bytesPosted > res.bytesLimit {
		return false, exception2.NewProtocolException("Post data exceed content-length: %d/%d", res.bytesPosted, res.bytesLimit)
	}

	oldAvailable := res.available
	if !res.bufferAvailable() {
		res.available = false
	}
	if oldAvailable && !res.available {
		baylog.Debug("%s response unavailable (_ _): posted=%d consumed=%d", res, res.bytesPosted, res.bytesConsumed)
	}

	return res.available, nil
}

/**
 * This method notifies the client that the response has ended.
 * Whether this process is synchronous or asynchronous is uncertain.
 * If it occurs synchronously, the tour instance will be disposed, and no further processing on the tour will be allowed
 */

func (res *TourResImpl) EndResContent(checkId int) exception.IOException {
	res.tour.CheckTourId(checkId)

	baylog.Debug("%s end ResContent", res)

	if res.tour.IsEnded() {
		baylog.Debug("%s Tour is already ended (Ignore).", res)
		return nil
	}

	if !res.tour.IsZombie() && res.tour.city != nil {
		res.tour.city.Log(res.tour)
	}

	// send end message
	if res.canCompress {

	}

	tourReturned := false
	lis := func() {
		baylog.Debug("CALLBACK!")
		res.tour.CheckTourId(checkId)
		res.tour.ship.ReturnTour(res.tour)
		tourReturned = true
	}

	var ioerr exception.IOException = nil
	for { // try catch
		if res.tour.IsZombie() || res.tour.IsAborted() {
			// Don't send peer any data. Do nothing
			baylog.Debug("%s Aborted or zombie tour. do nothing: %s state=%d", res, res.tour, res.tour.state)
			lis()

		} else {
			ioerr = res.tour.ship.SendEndTour(res.tour.ShipId(), res.tour, lis)

			if ioerr != nil {
				baylog.Debug("%s Error on sending end tour", res)
				lis()
				break
			}
		}

		break
	}

	// If tour is returned, we cannot change its state because
	// it will become uninitialized.
	baylog.Debug("%s Is the tour returned? : %v", res, tourReturned)
	if !tourReturned {
		res.tour.ChangeState(checkId, STATE_ENDED)
	}

	return ioerr
}

/**
 * This method sends an HTTP error response to the client.
 * Whether this process is carried out synchronously or asynchronously is uncertain
 */

func (res *TourResImpl) SendError(checkId int, status int, message string, err exception.Exception) exception.IOException {
	res.tour.CheckTourId(checkId)

	baylog.Debug("%s send error: status=%d, message=%s ex=%v", res, status, message, err)
	if err != nil {
		baylog.DebugE(err, "")
	}

	if res.tour.IsZombie() {
		return nil
	}

	if res.headerSent {
		baylog.Debug("Try to send error after response header is sent (Ignore)")
		baylog.Debug("%s: status=%d, message=%s", res, status, message)
		if err != nil {
			baylog.ErrorE(err, "")
		}

	} else {
		res.SetConsumeListener(tour.DevNullContentConsumeListener)

		if res.tour.IsZombie() || res.tour.IsAborted() {
			// Don't send peer any data. Do nothing
			baylog.Debug("%s Aborted or zombie tour. do nothing: %s state=%s", res, res.tour, res.tour.state)

		} else {
			// Create body
			desc := httpstatus.GetDescription(status)

			// print status
			body := fmt.Sprintf("<H1>%d %s</h1>\r\n", status, desc)

			res.headers.SetStatus(status)

			ioerr := res.sendErrorContent(body)
			if ioerr != nil {
				baylog.DebugE(ioerr, "%s Error in sending error", res)
				res.tour.ChangeState(checkId, STATE_ABORTED)
			}
			res.headerSent = true
		}
	}

	return res.EndResContent(checkId)
}

func (res *TourResImpl) SendHttpException(checkId int, hterr exception2.HttpException) exception.IOException {
	var ioerr exception.IOException = nil

	if hterr.Status() == httpstatus.MOVED_TEMPORARILY || hterr.Status() == httpstatus.MOVED_PERMANENTLY {
		ioerr = res.sendRedirect(checkId, hterr.Status(), hterr.Location())

	} else {
		ioerr = res.SendError(checkId, hterr.Status(), hterr.Error(), hterr)
	}

	if ioerr != nil {
		return ioerr
	}

	return nil
}

func (res *TourResImpl) SetConsumeListener(lis tour.ContentConsumeListener) {
	res.resConsumeListener = lis
	res.bytesConsumed = 0
	res.bytesPosted = 0
	res.available = true
}

func (res *TourResImpl) DetachConsumeListener() {
	res.resConsumeListener = nil
}

func (res *TourResImpl) BytesPosted() int {
	return res.bytesPosted
}

func (res *TourResImpl) BytesLimit() int {
	return res.bytesLimit
}

/****************************************/
/* Private functions                    */
/****************************************/

func (res *TourResImpl) sendRedirect(checkId int, status int, location string) exception.IOException {
	res.tour.CheckTourId(checkId)

	var ioerr exception.IOException = nil
	for { // try catch
		if res.headerSent {
			baylog.Error("Try to redirect after response header is sent (Ignore)")

		} else {
			res.SetConsumeListener(tour.DevNullContentConsumeListener)

			res.headers.SetStatus(status)
			res.headers.Set(headers.LOCATION, location)

			body := "<H2>Document Moved.</H2><BR>" + "<A HREF=\"" +
				location + "\">" + location + "</A>"

			ioerr = res.sendErrorContent(body)
			res.headerSent = true

			if ioerr != nil {
				break
			}
		}
		break
	}

	if ioerr != nil {
		res.tour.ChangeState(TOUR_ID_NOCHECK, STATE_ABORTED)
	}

	ioerr2 := res.EndResContent(checkId)
	if ioerr2 != nil {
		return ioerr2
	}

	return ioerr
}

func (res *TourResImpl) sendErrorContent(content string) exception.IOException {

	// Set content type
	if res.charset != "" {
		res.headers.SetContentType("text/html; charset=" + res.charset)
	} else {
		res.headers.SetContentType("text/html")
	}

	var ioerr exception.IOException = nil
	for { // try/catch
		var bytes []byte = nil
		if content != "" {
			// Create writer
			var encoding encoding.Encoding = nil
			if res.charset != "" {
				encoding = charutil.GetEncoding(res.charset)
			}

			if encoding != nil {
				reader := transform.NewReader(bytes2.NewReader([]byte(content)), encoding.NewEncoder())
				var err error
				bytes, err = io.ReadAll(reader)
				if err != nil {
					ioerr = exception.NewIOExceptionFromError(err)
					break
				}

			} else {
				bytes = []byte(content)
			}

			res.headers.SetContentLength(len(bytes))
		}

		ioerr = res.SendHeaders(res.tour.tourId)
		if ioerr != nil {
			break
		}

		if bytes != nil {
			_, ioerr = res.SendResContent(res.tour.tourId, bytes, 0, len(bytes))
			if ioerr != nil {
				break
			}
		}

		break
	}

	if ioerr != nil {
		return ioerr
	}

	return nil
}

/**
 * This method is called back when a part of the response data is actually sent to the client.
 * In this method, the internal buffer space is increased
 */

func (res *TourResImpl) consumed(checkId int, length int) {
	res.tour.CheckTourId(checkId)
	res.bytesConsumed += length

	baylog.Debug("%s resConsumed: len=%d posted=%d consumed=%d limit=%d",
		res.tour, length, res.bytesPosted, res.bytesConsumed, res.bytesLimit)

	resume := false
	oldAvailable := res.available
	if res.bufferAvailable() {
		res.available = true
	}
	if !oldAvailable && res.available {
		baylog.Debug("%s response available (^o^): posted=%d consumed=%d", res, res.bytesPosted, res.bytesConsumed)
		resume = true
	}

	if !res.tour.IsZombie() {
		if res.resConsumeListener == nil {
			baylog.Debug("Consume listener is null, so can not invoke callback")
		} else {
			res.resConsumeListener(length, resume)
		}
	}
}

func (res *TourResImpl) bufferAvailable() bool {
	return res.bytesPosted-res.bytesConsumed < bayserver.Harbor().TourBufferSize()
}
