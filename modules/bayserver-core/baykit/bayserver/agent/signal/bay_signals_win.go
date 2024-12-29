//go:build windows

package signal

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"os"
	"syscall"
)

var signalMap = map[os.Signal]string{
	os.Interrupt:    SIGN_AGENT_COMMAND_MEM_USAGE,
	syscall.SIGTERM: SIGN_AGENT_COMMAND_SHUTDOWN,
}

type BaySignals struct {
}

func NewBaySignals() *BaySignals {
	return &BaySignals{}
}

func (s *BaySignals) availableSignals() []os.Signal {
	sigs := make([]os.Signal, 0)
	for sig := range signalMap {
		sigs = append(sigs, sig)
	}
	return sigs
}

func (s *BaySignals) getCommand(sig os.Signal) string {
	return signalMap[sig]
}

func (s *BaySignals) getSignal(cmd string) (os.Signal, bool) {
	for sig, cmdName := range signalMap {
		if cmdName == cmd {
			return sig, true
		}
	}

	return nil, false
}

func (s *BaySignals) kill(pid int, sig os.Signal) exception.IOException {
	process, err := os.FindProcess(pid)
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	}

	err = process.Signal(sig)
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	}

	return nil
}
