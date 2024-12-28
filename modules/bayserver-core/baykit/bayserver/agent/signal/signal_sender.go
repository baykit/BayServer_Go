package signal

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/bcf/impl"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/docker/builtin"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bufio"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type SignalSender struct {
	bayPort int
	pidFile string
}

func NewSignalSender() *SignalSender {
	return &SignalSender{
		bayPort: builtin.DEFAULT_CONTROL_PORT,
		pidFile: builtin.DEFAULT_PID_FILE,
	}
}

/**
 * Send running BayServer a command
 */

func (s *SignalSender) SendCommand(cmd string) exception.IOException {
	perr := s.parseBayPort(bayserver.BservPlan())
	if perr != nil {
		baylog.ErrorE(perr, "")
		return exception.NewIOException(perr.Error())
	}

	var ioerr exception.IOException = nil
	for {
		// try catch
		if s.bayPort < 0 {
			var pid int
			pid, ioerr = s.readPidFile()
			if ioerr != nil {
				break
			}

			if pid <= 0 {
				return exception.NewIOException("Invalid process ID: %d", pid)
			}

			sig := getSignalFromCommand(cmd)
			if sig == 0 {
				return exception.NewIOException("Invalid command: %s", cmd)
			}

			ioerr = s.kill(pid, sig)
			if ioerr != nil {
				break
			}

		} else {
			baylog.Debug(baymessage.Get(symbol.MSG_SENDING_COMMAND, cmd))
			ioerr = s.send("localhost", s.bayPort, cmd)
			if ioerr != nil {
				break
			}
		}

		break
	}
	return nil
}

/**
 * Parse plan file and get port number of SignalAgent
 */

func (s *SignalSender) parseBayPort(plan string) bcf.ParseException {
	p := impl.NewBcfParser()
	doc, err := p.Parse(plan)
	if err != nil {
		return err
	}

	for _, o := range doc.ContentList {
		if elm, ok := o.(*bcf.BcfElement); ok {
			if strings.ToLower(elm.Name) == "harbor" {
				for _, o2 := range elm.ContentList {
					if kv, ok := o2.(*bcf.BcfKeyVal); ok {
						if strings.ToLower(kv.Key) == "controlport" {
							s.bayPort, _ = strconv.Atoi(kv.Value)

						} else if strings.ToLower(kv.Key) == "pidfile" {
							s.pidFile = kv.Value
						}
					}

				}
			}
		}
	}

	return nil
}

/**
 * Send another BayServer running host:port a command
 */

func (s *SignalSender) send(host string, port int, cmd string) exception.IOException {
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

func (s *SignalSender) kill(pid int, sig syscall.Signal) exception.IOException {
	baylog.Debug("Send signal %d to process %d", sig, pid)
	err := syscall.Kill(pid, sig)
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	}
	return nil
}

func (s *SignalSender) readPidFile() (int, exception.IOException) {

	var ioerr exception.IOException = nil
	for {
		// try catch
		file, err := os.Open(bayserver.GetLocation(s.pidFile))
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(err)
			break
		}
		defer file.Close()

		var bytes []byte
		bytes, err = io.ReadAll(file)
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(err)
			break
		}

		var num int
		content := string(bytes)
		num, err = strconv.Atoi(content)
		if err != nil {
			ioerr = exception.NewIOExceptionFromError(err)
			break
		}

		return num, nil
	}

	return -1, ioerr
}
