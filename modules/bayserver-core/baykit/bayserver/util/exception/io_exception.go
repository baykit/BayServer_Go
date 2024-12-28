package exception

type IOException interface {
	Exception
	IOError() string
}

type IOExceptionImpl struct {
	ExceptionImpl
}

func (i *IOExceptionImpl) IOError() string {
	return i.Error()
}

func NewIOException(format string, args ...interface{}) IOException {
	ex := &IOExceptionImpl{}
	ex.ConstructException(4, nil, format, args...)

	// interface check
	var _ Exception = ex
	var _ IOException = ex
	return ex
}

func NewIOExceptionFromError(err error) IOException {
	ex := IOExceptionImpl{}
	ex.ConstructException(4, err, err.Error())
	return &ex
}

func NewIOExceptionFromErrorWithMessage(err error, msg string) IOException {
	ex := IOExceptionImpl{}
	ex.ConstructException(4, err, msg)
	return &ex
}
