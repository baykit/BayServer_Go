package multiplexer

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/baylog"
	"sync"
)

type MultiplexerBase struct {
	agent       agent.GrandAgent
	rudderCount int
	rudders     map[rudder.Rudder]*common.RudderState
	lock        sync.Mutex
}

func NewMultiplexerBase(agt agent.GrandAgent) *MultiplexerBase {
	h := MultiplexerBase{
		agent:       agt,
		rudderCount: 0,
		rudders:     make(map[rudder.Rudder]*common.RudderState),
	}
	return &h
}

func (h *MultiplexerBase) String() string {
	return h.agent.String()
}

/****************************************/
/* Implements NetMultiplexer            */
/****************************************/

func (h *MultiplexerBase) AddRudderState(mpx common.Multiplexer, rd rudder.Rudder, st *common.RudderState) {
	st.Multiplexer = mpx
	h.lock.Lock()
	h.rudders[rd] = st
	h.lock.Unlock()
	h.rudderCount++
	st.Access()
}

func (h *MultiplexerBase) RemoveRudderState(rd rudder.Rudder) {
	h.lock.Lock()
	delete(h.rudders, rd)
	h.lock.Unlock()
}

func (h *MultiplexerBase) GetTransporter(rd rudder.Rudder) common.Transporter {
	return h.FindRudderState(rd).Transporter
}

func (h *MultiplexerBase) ConsumeOldestUnit(st *common.RudderState) bool {
	var consumed bool
	st.WriteLock.Lock()
	for { // try-catch
		if len(st.WriteQueue) == 0 {
			consumed = false
			break
		}
		u := st.WriteQueue[0]
		st.WriteQueue = arrayutil.RemoveAt(st.WriteQueue, 0)
		u.Done()
		consumed = true
		break
	}
	st.WriteLock.Unlock()

	return consumed
}

func (h *MultiplexerBase) CloseRudder(st *common.RudderState) {
	baylog.Debug("%s Close rd=%s", h.agent, st.Rudder)

	h.RemoveRudderState(st.Rudder)
	ioerr := st.Rudder.Close()
	if ioerr != nil {
		baylog.ErrorE(ioerr, "OS Close Error")
	}
}

func (h *MultiplexerBase) IsBusy() bool {
	return h.rudderCount >= h.agent.MaxInboundShips()
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (h *MultiplexerBase) FindRudderState(rd rudder.Rudder) *common.RudderState {
	h.lock.Lock()
	st := h.rudders[rd]
	h.lock.Unlock()
	return st
}

func (h *MultiplexerBase) closeAll() {
	for _, st := range h.rudders {
		if st.Rudder == h.agent.CommandReceiver().Rudder() {
			continue
		}
		h.CloseRudder(st)
	}
}
