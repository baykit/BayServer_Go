package bcf

import (
	"bayserver-core/baykit/bayserver/common/exception"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

type ParseException interface {
	exception.ConfigException
}

type ParseExceptionImpl struct {
	exception.ConfigExceptionImpl
}

func NewParseException(file string, line int, format string, args ...interface{}) ParseException {
	ex := &ParseExceptionImpl{}
	ex.ConstructException(4, nil, format, args...)

	var _ exception.ConfigException = ex
	var _ exception.BayException = ex
	var _ exception2.Exception = ex
	return ex
}
