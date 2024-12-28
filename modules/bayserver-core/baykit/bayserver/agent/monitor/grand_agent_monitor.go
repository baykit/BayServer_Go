package monitor

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/ioutil"
	"bayserver-core/baykit/bayserver/util/maputil"
	"os"
	"strconv"
	"time"
)

var numAgents = 0
var curId = 0

var monitors = map[int]*GrandAgentMonitor{}
var finale bool

type GrandAgentMonitor struct {
	agentId    int
	comSocket  rudder.Rudder
	anchorable bool
}

func NewGrandAgentMonitor(
	agentId int,
	anchorable bool,
	comSocket rudder.Rudder) *GrandAgentMonitor {

	return &GrandAgentMonitor{
		agentId:    agentId,
		comSocket:  comSocket,
		anchorable: anchorable,
	}
}

func (gm *GrandAgentMonitor) String() string {
	return "mon#" + strconv.Itoa(gm.agentId)
}

/****************************************/
/* Implements GrandAgentMonitor         */
/****************************************/

func (gm *GrandAgentMonitor) Start() {

	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		var ioerr exception.IOException = nil
		for { // loop

			var res int32
			res, ioerr = ioutil.ReadInt32(gm.comSocket)
			if ioerr != nil {
				break
			}
			if res == agent.CMD_CLOSE {
				baylog.Debug("%s read Close", gm)
				break

			} else {
				baylog.Debug("%s read OK: %d", gm, res)
			}
		}

		if ioerr != nil {
			baylog.ErrorE(ioerr, "Read from CommandReceiver error")
		}
		agentAborted(gm.agentId, gm.anchorable)
	}()
}

func (gm *GrandAgentMonitor) Shutdown() exception.IOException {
	baylog.Debug("%s send shutdown command", gm)
	return gm.send(agent.CMD_SHUTDOWN)
}

func (gm *GrandAgentMonitor) ReloadCert() exception.IOException {
	baylog.Debug("%s send reload command", gm)
	return gm.send(agent.CMD_RELOAD_CERT)
}

func (gm *GrandAgentMonitor) PrintUsage() exception.IOException {
	baylog.Debug("%s send usage command", gm)
	ioerr := gm.send(agent.CMD_MEM_USAGE)
	time.Sleep(1 * time.Second)

	return ioerr
}

/****************************************/
/* Private methods                      */
/****************************************/

func (gm *GrandAgentMonitor) send(cmd int32) exception.IOException {
	baylog.Debug("%s send command %d pipe=%s", gm, cmd, gm.comSocket)
	return ioutil.WriteInt32(gm.comSocket, cmd)
}

func (gm *GrandAgentMonitor) close() {
	ioerr := gm.comSocket.Close()
	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
	}
}

/****************************************/
/* Static functions                     */
/****************************************/

func GrandAgentMonitorInit(
	nAgents int) exception.IOException {

	numAgents = nAgents

	for i := 0; i < numAgents; i++ {
		err := grandAgentMonitorAdd(true)
		if err != nil {
			return err
		}
	}

	return nil
}

/**
 * Reload certificate for all agents
 */

func ReloadCertAll() exception.IOException {
	baylog.Debug("Reload All")
	for _, mon := range monitors {
		ioerr := mon.ReloadCert()
		if ioerr != nil {
			return ioerr
		}
	}
	return nil
}

/**
 * Restart all agents
 */

func RestartAll() exception.IOException {
	baylog.Debug("Restart All")
	oldMon := maputil.CopyMap(monitors)
	for _, mon := range oldMon {
		ioerr := mon.ReloadCert()
		if ioerr != nil {
			return ioerr
		}
	}
	return nil
}

func ShutdownAll() exception.IOException {
	baylog.Debug("Shutdown All")
	finale = true
	oldMon := maputil.CopyMap(monitors)
	for _, mon := range oldMon {
		ioerr := mon.Shutdown()
		if ioerr != nil {
			return ioerr
		}
	}
	return nil
}

func PrintUsageAll() {
	// print memory usage
	baylog.Info("BayServer MemUsage")

	for _, mon := range monitors {
		ioerr := mon.PrintUsage()
		if ioerr != nil {
			baylog.ErrorE(ioerr, "")
		}
	}
}

func grandAgentMonitorAdd(anchorable bool) exception.IOException {
	curId++
	agtId := curId
	if agtId > 100 {
		baylog.Error("Too many agents started")
		os.Exit(1)
	}

	var ioerr exception.IOException = nil
	for { // try catch
		var sockets []rudder.Rudder
		sockets, ioerr = ioutil.OpenLocalPipe()
		if ioerr != nil {
			break
		}

		agt := agent.Add(agtId, anchorable)
		agt.AddCommandReceiver(sockets[0])

		mon := NewGrandAgentMonitor(agtId, anchorable, sockets[1])
		monitors[agtId] = mon
		mon.Start()

		agt.Start()
		break
	}

	return ioerr

}

func agentAborted(agtId int, anchorable bool) {
	baylog.Error(baymessage.Get(symbol.MSG_GRAND_AGENT_SHUTDOWN, agtId))

	delete(monitors, agtId)

	if !finale {
		if len(monitors) < numAgents {
			ioerr := grandAgentMonitorAdd(anchorable)
			if ioerr != nil {
				baylog.ErrorE(ioerr, "")
			}
		}
	}
}
