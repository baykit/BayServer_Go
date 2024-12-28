package tour

import (
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/util/exception"
)

type ReqContentHandler interface {
	OnReadReqContent(tour Tour, buf []byte, start int, length int, lis ContentConsumeListener) exception.IOException
	OnEndReqContent(tour Tour) (exception.IOException, exception2.HttpException)
	OnAbortReq(tour Tour) bool
}

type DevNullContentHandler struct {
}

func NewDevNullContentHandler() *DevNullContentHandler {
	return &DevNullContentHandler{}
}

func (h *DevNullContentHandler) OnReadReqContent(
	tour Tour,
	buf []byte,
	start int,
	length int,
	lis ContentConsumeListener) exception.IOException {
	return nil
}

func (h *DevNullContentHandler) OnEndReqContent(tour Tour) (exception.IOException, exception2.HttpException) {
	return nil, nil
}

func (h *DevNullContentHandler) OnAbortReq(tour Tour) bool {
	return false
}
