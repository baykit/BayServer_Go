package signal

import (
	"bayserver-core/baykit/bayserver/util/exception"
)

type SignalSignSender struct {
	pid int
}

func NewSignalSignalSender(pid int) *SignalSignSender {
	return &SignalSignSender{pid: pid}
}

/**
 * Send running BayServer a command
 */

func (s *SignalSignSender) SendCommand(cmd string) exception.IOException {

	var ioerr exception.IOException = nil
	for {
		// try catch
		if s.pid <= 0 {
			return exception.NewIOException("Invalid process ID: %d", s.pid)
		}

		sigs := NewBaySignals()
		sig, exists := sigs.getSignal(cmd)
		if !exists {
			return exception.NewIOException("Invalid command: %s", cmd)
		}

		ioerr = sigs.kill(s.pid, sig)
		if ioerr != nil {
			break
		}
		break
	}
	return ioerr
}
