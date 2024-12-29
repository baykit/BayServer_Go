//go:build unix

package signal

import (
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
)

var signalMap = map[unix.Signal]string{
	unix.SIGALRM: SIGN_AGENT_COMMAND_RELOAD_CERT,
	unix.SIGTRAP: SIGN_AGENT_COMMAND_MEM_USAGE,
	unix.SIGHUP:  SIGN_AGENT_COMMAND_RESTART_AGENTS,
	unix.SIGTERM: SIGN_AGENT_COMMAND_SHUTDOWN,
	unix.SIGABRT: SIGN_AGENT_COMMAND_ABORT,
}

type BaySignals struct {
}

func NewBaySignals() *BaySignals {
	return &BaySignals{}
}

func (s *BaySignals) availableSignals() []unix.Signal {
	sigs := make([]unix.Signal, 0)
	for sig := range signalMap {
		sigs = append(sigs, sig)
	}
	return sigs
}

func (s *BaySignals) getCommand(sig os.Signal) string {
	return signalMap[sig.(unix.Signal)]
}

func (s *BaySignals) getSignal(cmd string) (unix.Signal, bool) {
	for sig, cmdName := range signalMap {
		if cmdName == cmd {
			return sig, true
		}
	}

	return 0, false
}

func (s *BaySignals) kill(pid int, sig syscall.Signal) exception.IOException {
	baylog.Debug("Send signal %d to process %d", sig, pid)
	err := syscall.Kill(pid, sig)
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	}
	return nil
}
