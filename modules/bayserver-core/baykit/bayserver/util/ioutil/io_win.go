//go:build windows

package ioutil

import (
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
)

func OpenLocalPipe() ([]rudder.Rudder, exception.IOException) {
	var ioerr exception.IOException = nil
	var lis net.Listener = nil
	var con1 net.Conn = nil
	var con2 net.Conn = nil

	for { // try catch
		var err error
		lis, err = net.Listen("tcp", "127.0.0.1:")
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(err)
			break
		}

		con1, err = net.Dial("tcp", lis.Addr().String())
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(err)
			break
		}

		con2, err = lis.Accept()
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(err)
			break
		}

		break
	}

	if lis != nil {
		_ = lis.Close()
	}
	/*
		if con1 != nil {
			_ = con1.Close()
		}
		if con2 != nil {
			_ = con2.Close()
		}
	*/
	if ioerr != nil {
		return nil, ioerr
	}

	return []rudder.Rudder{impl.NewTcpConnRudder(con1.(*net.TCPConn)), impl.NewTcpConnRudder(con2.(*net.TCPConn))}, nil
}
