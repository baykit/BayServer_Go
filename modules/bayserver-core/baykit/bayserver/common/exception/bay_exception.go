package exception

import (
	"bayserver-core/baykit/bayserver/util/exception"
)

type BayException interface {
	exception.Exception
}

type BayExceptionImpl struct {
	exception.ExceptionImpl
}

func NewBayException(format string, args ...interface{}) BayException {
	ex := &BayExceptionImpl{}
	ex.ConstructException(4, nil, format, args...)

	var _ exception.Exception = ex // interface check
	return ex
}

func NewBayExceptionFromError(err error) BayException {
	ex := BayExceptionImpl{}
	ex.ConstructException(4, err, err.Error())
	return &ex
}
