package signal

import (
	agent "bayserver-core/baykit/bayserver/agent/monitor"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"os"
	"strings"
)

type SignAgent interface {
	Run()
}

func SignalAgentInit(port int) {
	var agt SignAgent = nil
	if port > 0 {
		agt = NewTcpSignAgent(port)
	} else {
		agt = NewSignalSignAgent()
	}
	agt.Run()
}

/****************************************/
/* Static functions                     */
/****************************************/

func handleCommand(cmd string) {
	baylog.Debug("handle command: %s", cmd)
	var ioerr exception.IOException = nil
	for { // try-catch
		switch strings.ToLower(cmd) {
		case SIGN_AGENT_COMMAND_RELOAD_CERT:
			ioerr = agent.ReloadCertAll()

		case SIGN_AGENT_COMMAND_MEM_USAGE:
			agent.PrintUsageAll()

		case SIGN_AGENT_COMMAND_RESTART_AGENTS:
			ioerr = agent.RestartAll()
			if ioerr != nil {
				baylog.ErrorE(ioerr, "")
			}

		case SIGN_AGENT_COMMAND_SHUTDOWN:
			ioerr = agent.ShutdownAll()

		case SIGN_AGENT_COMMAND_ABORT:
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
