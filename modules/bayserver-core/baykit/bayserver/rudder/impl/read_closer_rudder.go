package impl

import (
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util/exception"
	"io"
)

type ReadCloserRudder struct {
	readCloser io.ReadCloser
}

func NewReadCloserRudder(rc io.ReadCloser) *ReadCloserRudder {
	rd := ReadCloserRudder{
		readCloser: rc,
	}
	return &rd
}

func (rd *ReadCloserRudder) String() string {
	return "ReadCloser"
}

func GetReadCloser(rd rudder.Rudder) io.ReadCloser {
	return rd.(*ReadCloserRudder).readCloser
}

/****************************************/
/* Implements Rudder                    */
/****************************************/

func (rd *ReadCloserRudder) Key() interface{} {
	return rd.readCloser
}

func (rd *ReadCloserRudder) Fd() int {
	return -1
}

func (rd *ReadCloserRudder) SetNonBlocking() exception.IOException {
	return nil
}

func (rd *ReadCloserRudder) Read(buf []byte) (int, exception.IOException) {
	n, err := rd.readCloser.Read(buf)
	if err != nil {
		if err.Error() == "EOF" {
			return 0, nil
		} else {
			return n, exception.NewIOExceptionFromError(err)
		}
	} else {
		return n, nil
	}
}

func (rd *ReadCloserRudder) Write(buf []byte) (int, exception.IOException) {
	return -1, exception.NewIOException("Not supported")
}

func (rd *ReadCloserRudder) Close() exception.IOException {
	err := rd.readCloser.Close()
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	} else {
		return nil
	}
}
