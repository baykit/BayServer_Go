package signal

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bufio"
	"net"
	"strconv"
	"strings"
	"time"
)

type TcpSignalAgent struct {
	port int
}

func NewTcpSignAgent(port int) *TcpSignalAgent {
	return &TcpSignalAgent{port: port}
}

/****************************************/
/* Static functions                     */
/****************************************/

func (agt *TcpSignalAgent) Run() {
	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		baylog.Info(baymessage.Get(symbol.MSG_OPEN_CTL_PORT, agt.port))
		var ioerr exception.IOException = nil
		server, err := net.Listen("tcp", ":"+strconv.Itoa(agt.port))
		if err != nil {
			bayserver.FatalError(exception.NewIOExceptionFromError(err))
		}
		lis := server.(*net.TCPListener)

		for { // infinite loop
			var s *net.TCPConn
			s, err = lis.AcceptTCP()

			if err != nil {
				ioerr = exception.NewIOExceptionFromError(err)
				break
			}

			for { // try-catch
				err = s.SetDeadline(time.Now().Add(5 * time.Second))
				if err != nil {
					ioerr = exception.NewIOExceptionFromError(err)
					break
				}

				var line string
				line, err = bufio.NewReader(s).ReadString('\n')
				if err != nil {
					ioerr = exception.NewIOExceptionFromError(err)
					break
				}

				baylog.Info(baymessage.Get(symbol.MSG_COMMAND_RECEIVED, line))
				handleCommand(strings.TrimSpace(line))

				_, err = s.Write([]byte("OK\n"))
				if err != nil {
					ioerr = exception.NewIOExceptionFromError(err)
					break
				}

				break
			}

			if ioerr != nil {
				baylog.ErrorE(ioerr, "")
			}

			if s != nil {
				err = s.Close()
				if err != nil {
					baylog.ErrorE(exception.NewIOExceptionFromError(err), "")
				}
			}
		}

	}()
}
