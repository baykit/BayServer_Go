package exception

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
)

type HttpException interface {
	exception.Exception

	Status() int
	Location() string
}

type HttpExceptionStruct struct {
	BayExceptionImpl

	status   int
	location string
}

func (h *HttpExceptionStruct) Status() int {
	return h.status
}

func (h *HttpExceptionStruct) Location() string {
	return h.location
}

func NewHttpException(status int, format string, args ...interface{}) HttpException {
	ex := HttpExceptionStruct{}
	ex.ConstructException(4, nil, format, args...)
	ex.status = status
	return &ex
}

func NewMovedTemporarily(location string) HttpException {
	ex := HttpExceptionStruct{}
	ex.ConstructException(4, nil, "")
	ex.status = httpstatus.MOVED_PERMANENTLY
	ex.location = location
	return &ex
}
