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
	"io"
	"os"
	"strconv"
	"strings"
)

type SignSender interface {
	SendCommand(cmd string) exception.IOException
}

/**
 * Send running BayServer a command
 */

func SendSign(cmd string) exception.IOException {
	bayPort, pidFile, perr := parseBayPort(bayserver.BservPlan())
	if perr != nil {
		baylog.ErrorE(perr, "")
		return exception.NewIOException(perr.Error())
	}

	var sender SignSender
	var ioerr exception.IOException = nil
	for {
		// try catch
		baylog.Debug(baymessage.Get(symbol.MSG_SENDING_COMMAND, cmd))
		if bayPort < 0 {
			var pid int
			pid, ioerr = readPidFile(pidFile)
			if ioerr != nil {
				break
			}

			if pid <= 0 {
				ioerr = exception.NewIOException("Invalid process ID: %d", pid)
				break
			}

			sender = NewUnixSignalSender(pid)

		} else {
			sender = NewTcpSignalSender(bayPort)
		}

		ioerr = sender.SendCommand(cmd)
		break
	}
	return ioerr
}

/**
 * Parse plan file and get port number of SignalAgent
 */

func parseBayPort(plan string) (int, string, bcf.ParseException) {
	p := impl.NewBcfParser()
	doc, err := p.Parse(plan)
	if err != nil {
		return 0, "", err
	}

	var bayPort int = builtin.DEFAULT_CONTROL_PORT
	var pidFile string = builtin.DEFAULT_PID_FILE
	for _, o := range doc.ContentList {
		if elm, ok := o.(*bcf.BcfElement); ok {
			if strings.ToLower(elm.Name) == "harbor" {
				for _, o2 := range elm.ContentList {
					if kv, ok := o2.(*bcf.BcfKeyVal); ok {
						if strings.ToLower(kv.Key) == "controlport" {
							bayPort, _ = strconv.Atoi(kv.Value)

						} else if strings.ToLower(kv.Key) == "pidfile" {
							pidFile = kv.Value
						}
					}

				}
			}
		}
	}

	return bayPort, pidFile, nil
}

func readPidFile(pidFile string) (int, exception.IOException) {

	var ioerr exception.IOException = nil
	for {
		// try catch
		file, err := os.Open(bayserver.GetLocation(pidFile))
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
