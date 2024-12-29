package agent

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util/exception"
)

const CMD_OK = 0
const CMD_CLOSE = 1
const CMD_RELOAD_CERT = 2
const CMD_MEM_USAGE = 3
const CMD_SHUTDOWN = 4
const CMD_ABORT = 5
const CMD_CATCHUP = 6

type GrandAgent interface {
	String() string
	AgentId() int
	SelectTimeoutSec() int
	MaxInboundShips() int
	NetMultiplexer() common.Multiplexer
	JobMultiplexer() common.Multiplexer
	CommandReceiver() *CommandReceiver
	Aborted() bool

	Shutdown()
	Abort()
	ReloadCert()
	PrintUsage()
	AddTimerHandler(th common.TimerHandler)
	RemoveTimerHandler(th common.TimerHandler)
	Ring()
	AddCommandReceiver(rd rudder.Rudder)
	SendAcceptedLetter(st *common.RudderState, clientRd rudder.Rudder, wakeup bool)
	SendConnectedLetter(st *common.RudderState, wakeup bool)
	SendReadLetter(st *common.RudderState, n int, address string, wakeup bool)
	SendWroteLetter(st *common.RudderState, n int, akeup bool)
	SendClosedLetter(st *common.RudderState, wakeup bool)
	SendErrorLetter(st *common.RudderState, err exception.Exception, wakeup bool)
	Start()
	AddPostpone(pp Postpone)
	ReqCatchUp()
	CatchUp()
}

var Init func(agentIds []int, nShips int)
var Get func(agentId int) GrandAgent
var AddLifeCycleListener func(lis common.LifecycleListener)
var Add func(agentId int, anchorable bool) GrandAgent
