package signal

import (
	agent "bayserver-core/baykit/bayserver/agent/monitor"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bufio"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const SIGNAL_AGENT_COMMAND_RELOAD_CERT = "reloadcert"
const SIGNAL_AGENT_COMMAND_MEM_USAGE = "memusage"
const SIGNAL_AGENT_COMMAND_RESTART_AGENTS = "restartagents"
const SIGNAL_AGENT_COMMAND_SHUTDOWN = "shutdown"
const SIGNAL_AGENT_COMMAND_ABORT = "abort"

var commands = []string{
	SIGNAL_AGENT_COMMAND_RELOAD_CERT,
	SIGNAL_AGENT_COMMAND_MEM_USAGE,
	SIGNAL_AGENT_COMMAND_RESTART_AGENTS,
	SIGNAL_AGENT_COMMAND_SHUTDOWN,
	SIGNAL_AGENT_COMMAND_ABORT,
}

var signalMap = map[unix.Signal]string{}

func SignalAgentInit(port int) {
	if port > 0 {
		runSignalAgent(port)
	} else {
		for _, cmd := range commands {
			sig := getSignalFromCommand(cmd)

			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig)

			cmd2 := cmd
			go func() {
				defer func() {
					bayserver.BDefer()
				}()

				for {
					_ = <-ch
					handleCommand(cmd2)
				}
			}()
		}

	}
}

/****************************************/
/* Static functions                     */
/****************************************/

func runSignalAgent(port int) {
	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		var ioerr exception.IOException = nil
		server, err := net.Listen("tcp", ":"+strconv.Itoa(port))
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
				handleCommand(line)

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

func handleCommand(cmd string) {
	baylog.Debug("handle command: %s", cmd)
	var ioerr exception.IOException = nil
	for { // try-catch
		switch strings.ToLower(cmd) {
		case SIGNAL_AGENT_COMMAND_RELOAD_CERT:
			ioerr = agent.ReloadCertAll()

		case SIGNAL_AGENT_COMMAND_MEM_USAGE:
			agent.PrintUsageAll()

		case SIGNAL_AGENT_COMMAND_RESTART_AGENTS:
			ioerr = agent.RestartAll()
			if ioerr != nil {
				baylog.ErrorE(ioerr, "")
			}

		case SIGNAL_AGENT_COMMAND_SHUTDOWN:
			ioerr = agent.ShutdownAll()

		case SIGNAL_AGENT_COMMAND_ABORT:
			os.Exit(1)

		default:
			baylog.Error("Unknown command: " + cmd)
		}

		break
	}

	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
	}

}

func getSignalFromCommand(command string) unix.Signal {
	initSignalMap()
	for sig, cmd := range signalMap {
		if strings.ToLower(cmd) == strings.ToLower(command) {
			return sig
		}
	}

	return 0
}

func initSignalMap() {
	if len(signalMap) > 0 {
		return
	}

	signalMap[unix.SIGALRM] = SIGNAL_AGENT_COMMAND_RELOAD_CERT
	signalMap[unix.SIGTRAP] = SIGNAL_AGENT_COMMAND_MEM_USAGE
	signalMap[unix.SIGHUP] = SIGNAL_AGENT_COMMAND_RESTART_AGENTS
	signalMap[unix.SIGTERM] = SIGNAL_AGENT_COMMAND_SHUTDOWN
	signalMap[unix.SIGABRT] = SIGNAL_AGENT_COMMAND_ABORT

}
