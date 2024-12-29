package agent

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/memusage"
	ship "bayserver-core/baykit/bayserver/ship/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/ioutil"
	"encoding/binary"
)

type CommandReceiver struct {
	ship.ShipImpl
	Closed bool
}

func NewCommandReceiver() *CommandReceiver {
	res := CommandReceiver{
		Closed: false,
	}
	res.ShipImpl.Construct()
	return &res
}

/****************************************/
/* Implements Ship                      */
/****************************************/

func (c *CommandReceiver) NotifyHandshakeDone(protocol string) (common.NextSocketAction, exception.IOException) {
	//TODO implement me
	panic("implement me")
}

func (c *CommandReceiver) NotifyConnect() (common.NextSocketAction, exception.IOException) {
	//TODO implement me
	panic("implement me")
}

func (c *CommandReceiver) NotifyRead(buf []byte) (common.NextSocketAction, exception.IOException) {
	var num uint32
	num = binary.BigEndian.Uint32(buf)
	c.onReadCommand(num)
	return common.NEXT_SOCKET_ACTION_CONTINUE, nil
}

func (c *CommandReceiver) NotifyEof() common.NextSocketAction {
	return -1
}

func (c *CommandReceiver) NotifyError(e exception.Exception) {
	baylog.ErrorE(e, "")
}

func (c *CommandReceiver) NotifyProtocolError(e exception2.ProtocolException) (bool, exception.IOException) {
	bayserver.FatalError(exception.NewSink(""))
	return false, nil
}

func (c *CommandReceiver) NotifyClose() {
}

func (c *CommandReceiver) CheckTimeout(durationSec int) bool {
	return false
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (c *CommandReceiver) onReadCommand(cmd uint32) {
	var ioerr exception.IOException
	var agent = Get(c.AgentId())

catch:
	for { // Try catch

		baylog.Debug("%s receive command %d pipe=%d", agent, cmd, c.Rudder())
		switch cmd {
		case CMD_RELOAD_CERT:
			agent.ReloadCert()

		case CMD_MEM_USAGE:
			memusage.Get(agent.AgentId()).PrintUsage(0)

		case CMD_SHUTDOWN:
			agent.Shutdown()

		case CMD_ABORT:
			ioerr = c.SendCommandToMonitor(agent, CMD_OK, true)
			if ioerr != nil {
				break
			}
			agent.Abort()
			break catch

		case CMD_CATCHUP:
			agent.CatchUp()
			break catch

		default:
			baylog.Error("Unknown command: %d", cmd)
		}

		if ioerr != nil {
			break
		}

		ioerr = c.SendCommandToMonitor(agent, CMD_OK, false)

		if cmd == CMD_SHUTDOWN {
			c.End()
		}
		break
	}

	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
		c.close()
	}
	baylog.Info("%s Command ended", agent)
}

func (c *CommandReceiver) SendCommandToMonitor(agt GrandAgent, cmd int32, sync bool) exception.IOException {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(cmd))
	if sync {
		return ioutil.WriteInt32(c.Rudder(), cmd)

	} else {
		return agt.NetMultiplexer().ReqWrite(c.Rudder(), bytes, nil, nil, nil)
	}
}

func (c *CommandReceiver) End() {
	baylog.Debug("%s Command Receiver -> Send end", c)
	ioerr := ioutil.WriteInt32(c.Rudder(), CMD_CLOSE)
	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
	}
	c.close()
}

func (c *CommandReceiver) close() {
	ioerr := c.Rudder().Close()
	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
	}
	c.Closed = true
}
