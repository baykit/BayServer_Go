package exception

type Sink interface {
	Exception
}

type SinkStruct struct {
	ExceptionImpl
	dummySink int
}

func NewSink(format string, args ...interface{}) Sink {
	ex := SinkStruct{}
	ex.ConstructException(4, nil, format, args...)
	return &ex
}

func NewSinkFromError(err error) Sink {
	ex := SinkStruct{}
	ex.ConstructException(4, err, err.Error())
	return &ex
}
