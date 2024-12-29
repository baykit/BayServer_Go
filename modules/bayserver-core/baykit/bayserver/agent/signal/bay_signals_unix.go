//go:build unix

package signal

import (
	"golang.org/x/sys/unix"
)

var signalMap = map[unix.Signal]string{
	unix.SIGALRM: SIGNAL_AGENT_COMMAND_RELOAD_CERT,
	unix.SIGTRAP: SIGNAL_AGENT_COMMAND_MEM_USAGE,
	unix.SIGHUP:  SIGNAL_AGENT_COMMAND_RESTART_AGENTS,
	unix.SIGTERM: SIGNAL_AGENT_COMMAND_SHUTDOWN,
	unix.SIGABRT: SIGNAL_AGENT_COMMAND_ABORT,
}

func (s *SignalSignSender) kill(pid int, sig syscall.Signal) exception.IOException {
	baylog.Debug("Send signal %d to process %d", sig, pid)
	err := syscall.Kill(pid, sig)
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	}
	return nil
}
