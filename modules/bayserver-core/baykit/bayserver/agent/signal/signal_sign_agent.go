package signal

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"os"
	"os/signal"
)

const SIGNAL_AGENT_COMMAND_RELOAD_CERT = "reloadcert"
const SIGNAL_AGENT_COMMAND_MEM_USAGE = "memusage"
const SIGNAL_AGENT_COMMAND_RESTART_AGENTS = "restartagents"
const SIGNAL_AGENT_COMMAND_SHUTDOWN = "shutdown"
const SIGNAL_AGENT_COMMAND_ABORT = "abort"

type SignalSignAgent struct {
}

func NewSignalSignAgent() *SignalSignAgent {
	return &SignalSignAgent{}
}

func (agt *SignalSignAgent) Run() {
	sigs := NewBaySignals()
	ch := make(chan os.Signal, 1)
	for _, sig := range sigs.availableSignals() {
		signal.Notify(ch, sig)
	}

	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		for {
			sig := <-ch
			cmd := sigs.getCommand(sig)
			handleCommand(cmd)
		}
	}()
}
