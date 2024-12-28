package impl

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
	"runtime"
)

type ListenerRudder struct {
	TcpListener *net.TCPListener
	fd          int
}

func NewListenerRudder(lis *net.TCPListener) *ListenerRudder {
	rd := ListenerRudder{
		TcpListener: lis,
	}
	if runtime.GOOS != "windows" {
		file, err := lis.File()
		if err != nil {
			baylog.ErrorE(exception.NewIOExceptionFromError(err), "")
		} else {
			rd.fd = int(file.Fd())
		}
	}
	return &rd
}

func (rd *ListenerRudder) String() string {
	return "Listener[" + rd.TcpListener.Addr().String() + "]"
}

func GetListener(rd rudder.Rudder) *net.TCPListener {
	return rd.(*ListenerRudder).TcpListener
}

/****************************************/
/* Implements Rudder                    */
/****************************************/

func (rd *ListenerRudder) Key() interface{} {
	return rd.TcpListener
}

func (rd *ListenerRudder) Fd() int {
	return rd.fd
}

func (rd *ListenerRudder) Read(buf []byte) (int, exception.IOException) {
	bayserver.FatalError(exception.NewSink(""))
	return 0, nil
}

func (rd *ListenerRudder) Write(buf []byte) (int, exception.IOException) {
	bayserver.FatalError(exception.NewSink(""))
	return 0, nil
}

func (rd *ListenerRudder) Close() exception.IOException {
	err := rd.TcpListener.Close()
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	} else {
		return nil
	}
}
