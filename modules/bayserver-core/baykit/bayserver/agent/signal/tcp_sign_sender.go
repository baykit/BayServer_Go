package signal

import (
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bufio"
	"net"
	"strconv"
)

type TcpSignalSender struct {
	port int
}

func NewTcpSignalSender(port int) *TcpSignalSender {
	return &TcpSignalSender{
		port: port,
	}
}

/**
 * Send running BayServer a command
 */

func (s *TcpSignalSender) SendCommand(cmd string) exception.IOException {
	var ioerr exception.IOException = nil

	baylog.Debug(baymessage.Get(symbol.MSG_SENDING_COMMAND, cmd))
	ioerr = s.send("localhost", s.port, cmd)
	return ioerr
}

/**
 * Send another BayServer running host:port a command
 */

func (s *TcpSignalSender) send(host string, port int, cmd string) exception.IOException {
	var ioerr exception.IOException = nil
	for {
		// try catch
		conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(err)
			break
		}

		defer conn.Close()

		w := bufio.NewWriter(conn)
		_, err = w.WriteString(cmd)
		_, err = w.WriteString("\n")
		err = w.Flush()
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(err)
			break
		}

		r := bufio.NewReader(conn)
		_, _, err = r.ReadLine()
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(err)
			break
		}

		break
	}

	if ioerr != nil {
		return ioerr
	}
	return nil
}
