package warpship

import (
	"bayserver-core/baykit/bayserver/agent"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"strconv"
)

type WarpData struct {
	warpShip   WarpShip
	warpShipId int
	WarpId     int
	reqHeaders *headers.Headers
	ResHeaders *headers.Headers
	Started    bool
	Ended      bool
}

func NewWarpData(wsip WarpShip, wid int) *WarpData {
	wdata := WarpData{
		warpShip:   wsip,
		warpShipId: wsip.ShipId(),
		WarpId:     wid,
		reqHeaders: headers.NewHeaders(),
		ResHeaders: headers.NewHeaders(),
	}
	var _ tour.ReqContentHandler = &wdata // cast check
	return &wdata
}

func (w *WarpData) String() string {
	return w.warpShip.String() + " wtur#" + strconv.Itoa(w.WarpId)
}

/****************************************/
/* Implements ReqCpntentHandler         */
/****************************************/

func (w *WarpData) OnReadReqContent(tur tour.Tour, buf []byte, start int, length int, lis tour.ContentConsumeListener) exception.IOException {
	baylog.Debug("%s onReadReqContent tur=%s len=%d", w.warpShip, tur, length)
	w.warpShip.CheckShipId(w.warpShipId)
	maxLen := w.warpShip.ProtocolHandler().MaxReqPacketDataSize()
	for pos := 0; pos < length; pos += maxLen {
		postLen := length - pos
		if postLen > maxLen {
			postLen = maxLen
		}
		turId := tur.TourId()

		if !w.Started {
			// The buffer will become corrupted due to reuse.
			buf = arrayutil.CopyArray(buf)
		}

		ioerr := w.warpShip.WarpHandler().SendContent(
			tur,
			buf,
			start+pos,
			postLen,
			func() { tur.Req().Consumed(turId, length, lis) })

		if ioerr != nil {
			return ioerr
		}
	}

	return nil
}

func (w *WarpData) OnEndReqContent(tur tour.Tour) (exception.IOException, exception2.HttpException) {
	baylog.Debug("%s endReqContent tur=%s", w.warpShip, tur)
	w.warpShip.CheckShipId(w.warpShipId)
	ioerr := w.warpShip.WarpHandler().SendEnd(
		tur,
		false,
		func() {
			agt := agent.Get(w.warpShip.AgentId())
			agt.NetMultiplexer().ReqRead(w.warpShip.Rudder())
		})

	return ioerr, nil
}

func (w *WarpData) OnAbortReq(tur tour.Tour) bool {
	baylog.Debug("%s onAbortReq tur=%s", w.warpShip, tur)
	w.warpShip.CheckShipId(w.warpShipId)
	w.warpShip.Abort(w.warpShipId)
	return false
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (w *WarpData) Start() exception.IOException {
	if !w.Started {
		baylog.Debug("%s Start Warp tour", w)
		ioerr := w.warpShip.Flush()
		if ioerr != nil {
			return ioerr
		}
		w.Started = true
	}

	return nil
}

/****************************************/
/* Static functions                     */
/****************************************/

func WarpDataGet(tur tour.Tour) *WarpData {
	return tur.Req().GetReqContentHandler().(*WarpData)
}
