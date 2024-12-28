//go:build unix

package impl

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
	"syscall"
)

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
			value, err = syscall.GetsockoptInt(int(fd), level, opt)
		})

		if err != nil {
			return 0, exception.NewIOExceptionFromError(err)
		}

		return value, nil

	} else {
		return 8192, nil
	}

}
