package exception

import "bayserver-core/baykit/bayserver/util/exception"

type ProtocolException interface {
	exception.IOException

	ProtocolExceptionDummy()
}

type ProtocolExceptionImpl struct {
	exception.IOExceptionImpl
}

// Dummy function
func (e *ProtocolExceptionImpl) ProtocolExceptionDummy() {
}

func NewProtocolException(format string, args ...interface{}) ProtocolException {
	ex := &ProtocolExceptionImpl{}
	ex.ConstructException(4, nil, format, args...)

	// interface check
	var _ exception.Exception = ex
	var _ exception.IOException = ex
	var _ ProtocolException = ex
	return ex
}

func NewProtocolExceptionFromError(err error) ProtocolException {
	ex := ProtocolExceptionImpl{}
	ex.ConstructException(4, err, err.Error())
	return &ex
}
