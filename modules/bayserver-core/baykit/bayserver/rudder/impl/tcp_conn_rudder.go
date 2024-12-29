package impl

import (
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"fmt"
	"io"
	"net"
	"syscall"
)

type TcpConnRudder struct {
	Conn net.Conn
}

func NewTcpConnRudder(con net.Conn) *TcpConnRudder {
	rd := TcpConnRudder{
		Conn: con,
	}
	var _ rudder.ConnRudder = &rd // cast check
	return &rd
}

func NewTcpConnRudderUnconnected() *TcpConnRudder {
	rd := TcpConnRudder{
		Conn: nil,
	}
	return &rd
}

func (rd *TcpConnRudder) String() string {
	return fmt.Sprintf("TcpConRudder[%p]", rd)
}

func GetConn(rd rudder.Rudder) net.Conn {
	return rd.(*TcpConnRudder).Conn
}

/****************************************/
/* Implements Rudder                    */
/****************************************/

func (rd *TcpConnRudder) Key() interface{} {
	return rd.Conn
}

func (rd *TcpConnRudder) Fd() int {
	return 0
}

func (rd *TcpConnRudder) Read(buf []byte) (int, exception.IOException) {
	n, err := rd.Conn.Read(buf)
	if err != nil {
		// "In a Windows environment, when a TCP connection ends,
		// the type of the error returned can be *errors.errorString,
		// and this error may contain the message 'EOF'."
		if err == io.EOF || err.Error() == "EOF" {
			baylog.Debug("EOF detected")
			return 0, nil
		}
		return 0, exception.NewIOExceptionFromError(err)
	} else {
		return n, nil
	}
}

func (rd *TcpConnRudder) Write(buf []byte) (int, exception.IOException) {
	n, err := rd.Conn.Write(buf)
	if err != nil {
		return 0, exception.NewIOExceptionFromError(err)
	} else {
		return n, nil
	}
}

func (rd *TcpConnRudder) Close() exception.IOException {
	var ioerr exception.IOException = nil
	for { // try-catch
		if rd.Conn != nil {
			err := rd.Conn.Close()
			if err != nil {
				ioerr = exception.NewIOExceptionFromError(err)
				break
			}
		}
		break
	}

	return ioerr
}

/****************************************/
/* Implements ConnRudder                */
/****************************************/

func (rd *TcpConnRudder) GetRemotePort() int {
	return rd.Conn.RemoteAddr().(*net.TCPAddr).Port
}

func (rd *TcpConnRudder) GetRemoteAddress() string {
	return rd.Conn.RemoteAddr().(*net.TCPAddr).IP.String()
}

func (rd *TcpConnRudder) GetLocalAddress() string {
	return rd.Conn.LocalAddr().(*net.TCPAddr).IP.String()
}

func (rd *TcpConnRudder) GetSocketReceiveBufferSize() (int, exception.IOException) {
	size, ioerr := getSockOpt(rd.Conn, syscall.SOL_SOCKET, syscall.SO_RCVBUF)
	if ioerr != nil {
		return -1, ioerr
	}

	return size, nil
}

func getSockOpt(conn net.Conn, level, opt int) (int, exception.IOException) {
	if tcpCon, ok := conn.(*net.TCPConn); ok {
		var value int
		var err error

		sysCon, err := tcpCon.SyscallConn()
		err = sysCon.Control(func(fd uintptr) {
			value, err = getSockOptInt(int(fd), level, opt)
		})

		if err != nil {
			return 0, exception.NewIOExceptionFromError(err)
		}

		return value, nil

	} else {
		return 8192, nil
	}

}
