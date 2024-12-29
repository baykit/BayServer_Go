package agent

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/agent/letter"
	"bayserver-core/baykit/bayserver/agent/multiplexer"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"strconv"
	"sync"
)

const SELECT_TIMEOUT_SEC = 10

var agentCount int
var maxShips int
var maxAgentId int
var agents = map[int]agent.GrandAgent{}
var listeners []common.LifecycleListener

type GrandAgentImpl struct {
	agentId           int
	anchorable        bool
	letterQueue       []letter.Letter
	timeoutSec        int
	maxInboundShips   int
	netMultiplexer    common.Multiplexer
	jobMultiplexer    common.Multiplexer
	recipient         common.Recipient
	aborted           bool
	timerHandlers     []common.TimerHandler
	letterQueueLock   sync.Mutex
	commandReceiver   *agent.CommandReceiver
	lastTimeoutCheck  int64
	postponeQueue     []agent.Postpone
	postponeQueueLock sync.Mutex
}

func NewGrandAgent(
	agentId int,
	maxShips int,
	anchorable bool) agent.GrandAgent {

	g := GrandAgentImpl{
		agentId:         agentId,
		letterQueue:     make([]letter.Letter, 0),
		timeoutSec:      SELECT_TIMEOUT_SEC,
		maxInboundShips: maxShips,
		aborted:         false,
		timerHandlers:   []common.TimerHandler{},
	}

	if g.maxInboundShips == 0 {
		g.maxInboundShips = 1
	}

	g.jobMultiplexer = multiplexer.NewJobMultiplexer(&g, true)
	g.anchorable = anchorable
	g.netMultiplexer = g.jobMultiplexer

	switch bayserver.Harbor().Recipient() {
	case docker.RECIPIENT_TYPE_SPIDER:
		bayserver.FatalError(exception.NewSink(""))

	case docker.RECIPIENT_TYPE_PIPE:
		g.recipient = multiplexer.NewPipeRecipient()
	}

	switch bayserver.Harbor().NetMultiplexer() {
	case docker.MULTI_PLEXER_TYPE_SPIDER:
		bayserver.FatalError(exception.NewSink(""))

	case docker.MULTI_PLEXER_TYPE_JOB:
		g.netMultiplexer = g.jobMultiplexer

	default:
		bayserver.FatalError(
			exception.NewSink(
				"Multiplexer not supported: %s",
				docker.GetMultiplexerTypeName(bayserver.Harbor().NetMultiplexer())))
	}
	//g.netMultiplexer = netMultiplexer.NewSelectMultiplexer(&g, true)
	return &g
}

func (g *GrandAgentImpl) String() string {
	return "agt#" + strconv.Itoa(g.agentId)
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (g *GrandAgentImpl) Run() {
	baylog.Info(baymessage.Get(symbol.MSG_RUNNING_GRAND_AGENT, g))

	g.netMultiplexer.ReqRead(g.commandReceiver.Rudder())

	if g.anchorable {
		for rd := range bayserver.AnchorablePortMap() {
			g.netMultiplexer.AddRudderState(rd, common.NewRudderState(rd, nil))
		}
	}

	var err exception.Exception = nil
	var busy = true

catch:
	for {
		testBusy := g.netMultiplexer.IsBusy()
		if testBusy != busy {
			busy = testBusy
			if busy {
				g.netMultiplexer.OnBusy()

			} else {
				g.netMultiplexer.OnFree()
			}
		}

		_, err = g.recipient.Receive(true)
		if err != nil {
			break catch
		}

		if g.aborted {
			baylog.Info("%s aborted by another thread", g)
			break catch
		}

		if len(g.letterQueue) == 0 {
			// timed out
			// check per 10 seconds
			if sysutil.CurrentTimeSecs()-g.lastTimeoutCheck >= 10 {
				g.Ring()
			}
		}

		for len(g.letterQueue) > 0 {
			var let letter.Letter
			g.letterQueueLock.Lock()
			let = g.letterQueue[0]
			g.letterQueue = arrayutil.RemoveAt(g.letterQueue, 0)
			g.letterQueueLock.Unlock()

			switch let2 := let.(type) {

			case *letter.AcceptedLetter:
				g.onAccepted(let2)

			case *letter.ConnectedLetter:
				g.onConnected(let2)

			case *letter.ReadLetter:
				g.onRead(let2)

			case *letter.WroteLetter:
				g.onWrote(let2)

			case *letter.ClosedLetter:
				g.onClosed(let2)

			case *letter.ErrorLetter:
				err = g.onError(let2)
			}

			if err != nil {
				break catch
			}
		}
	}

	if err != nil {
		bayserver.FatalError(err)

	} else {
		baylog.Info("%s end", g)
		g.Shutdown()
	}
}

func (g *GrandAgentImpl) Start() {
	go func() {
		bayserver.BDefer()
		g.Run()
	}()
}

func (g *GrandAgentImpl) AgentId() int {
	return g.agentId
}

func (g *GrandAgentImpl) SelectTimeoutSec() int {
	return g.timeoutSec
}

func (g *GrandAgentImpl) MaxInboundShips() int {
	return g.maxInboundShips
}

func (g *GrandAgentImpl) NetMultiplexer() common.Multiplexer {
	return g.netMultiplexer
}

func (g *GrandAgentImpl) JobMultiplexer() common.Multiplexer {
	return g.jobMultiplexer
}

func (g *GrandAgentImpl) CommandReceiver() *agent.CommandReceiver {
	return g.commandReceiver
}

func (g *GrandAgentImpl) Aborted() bool {
	return g.aborted
}

func (g *GrandAgentImpl) Shutdown() {
	baylog.Debug("%s shutdown aborted=%v", g, g.aborted)
	if g.aborted {
		return
	}
	g.aborted = true
	g.netMultiplexer.Shutdown()

	for _, lis := range listeners {
		lis.Remove(g.agentId)
	}

	delete(agents, g.agentId)
}

func (g *GrandAgentImpl) Abort() {
	baylog.Fatal("%s abort", g)
}

func (g *GrandAgentImpl) ReloadCert() {

}

func (g *GrandAgentImpl) PrintUsage() {

}

func (g *GrandAgentImpl) AddTimerHandler(th common.TimerHandler) {
	g.timerHandlers = append(g.timerHandlers, th)
}

func (g *GrandAgentImpl) RemoveTimerHandler(th common.TimerHandler) {
	g.timerHandlers, _ = arrayutil.RemoveObject(g.timerHandlers, th)
}

func (g *GrandAgentImpl) Ring() {
	for _, th := range g.timerHandlers {
		th.OnTimer()
	}
	g.lastTimeoutCheck = sysutil.CurrentTimeSecs()
}

func (g *GrandAgentImpl) AddCommandReceiver(rd rudder.Rudder) {
	g.commandReceiver = agent.NewCommandReceiver()
	comTp := multiplexer.NewPlainTransporter(g.netMultiplexer, g.commandReceiver, true, 8, false)
	g.commandReceiver.Init(g.agentId, rd, comTp)
	g.netMultiplexer.AddRudderState(g.commandReceiver.Rudder(), common.NewRudderState(g.commandReceiver.Rudder(), comTp))
}

func (g *GrandAgentImpl) SendAcceptedLetter(st *common.RudderState, clientRd rudder.Rudder, wakeup bool) {
	g.sendLetter(letter.NewAcceptedLetter(st, clientRd), wakeup)
}

func (g *GrandAgentImpl) SendConnectedLetter(st *common.RudderState, wakeup bool) {
	g.sendLetter(letter.NewConnectedLetter(st), wakeup)
}

func (g *GrandAgentImpl) SendReadLetter(st *common.RudderState, n int, address string, wakeup bool) {
	g.sendLetter(letter.NewReadLetter(st, n, address), wakeup)
}

func (g *GrandAgentImpl) SendWroteLetter(st *common.RudderState, n int, wakeup bool) {
	g.sendLetter(letter.NewWroteLetter(st, n), wakeup)
}

func (g *GrandAgentImpl) SendClosedLetter(st *common.RudderState, wakeup bool) {
	g.sendLetter(letter.NewClosedLetter(st), wakeup)
}

func (g *GrandAgentImpl) SendErrorLetter(st *common.RudderState, err exception.Exception, wakeup bool) {
	g.sendLetter(letter.NewErrorLetter(st, err), wakeup)
}

func (g *GrandAgentImpl) AddPostpone(pp agent.Postpone) {
	g.postponeQueue = append(g.postponeQueue, pp)
}

func (g *GrandAgentImpl) ReqCatchUp() {
	baylog.Debug("%s Req catchUp", g)
	if g.countPostpone() > 0 {
		g.CatchUp()

	} else {
		ioerr := g.commandReceiver.SendCommandToMonitor(g, agent.CMD_CATCHUP, false)
		if ioerr != nil {
			baylog.ErrorE(ioerr, "")
			g.Abort()
		}
	}
}

/****************************************/
/* private functions                    */
/****************************************/

func (g *GrandAgentImpl) sendLetter(let letter.Letter, wakeup bool) {
	g.letterQueueLock.Lock()
	g.letterQueue = append(g.letterQueue, let)
	g.letterQueueLock.Unlock()

	if wakeup {
		g.recipient.Wakeup()
	}
}

func (g *GrandAgentImpl) onAccepted(let *letter.AcceptedLetter) {
	baylog.Debug("%s onAccept client rd=%s", g, let.ClientRudder)

	st := let.State()
	p := bayserver.AnchorablePortMap()[st.Rudder]
	if p == nil {
		serverRudders := bayserver.AnchorablePortMap()
		baylog.Fatal("Rudder '%s' is not server rudder list: %s", st.Rudder, serverRudders)
		bayserver.FatalError(exception.NewSink(""))
	}

	herr := p.OnConnected(g.agentId, let.ClientRudder)

	if herr != nil {
		st.Transporter.OnError(st.Rudder, herr)
		g.nextAction(st, common.NEXT_SOCKET_ACTION_CLOSE, false)
	}

	if !g.netMultiplexer.IsBusy() {
		st.Multiplexer.(common.Multiplexer).NextAccept(st)
	}
}

func (g *GrandAgentImpl) onConnected(let *letter.ConnectedLetter) {
	st := let.State()
	if st.Closed {
		baylog.Debug("%s Rudder is already closed: rd=%s", g, st.Rudder)
		return
	}

	baylog.Debug("%s connected rd=%s", g, st.Rudder)

	nextAct, ioerr := st.Transporter.OnConnect(st.Rudder)
	baylog.Debug("%s nextAct=%d", g, nextAct)

	if ioerr != nil {
		st.Transporter.OnError(st.Rudder, ioerr)
		nextAct = common.NEXT_SOCKET_ACTION_CLOSE
	}

	if nextAct == common.NEXT_SOCKET_ACTION_READ {
		// Read more
		st.Multiplexer.CancelWrite(st)
	}

	g.nextAction(st, nextAct, false)
}

func (g *GrandAgentImpl) onRead(let *letter.ReadLetter) {
	st := let.State()

	var ioerr exception.IOException = nil
	var nextAct common.NextSocketAction
	for { // try-catch
		baylog.Debug("%s read %d bytes (rd=%s)", g, let.NBytes, st.Rudder)
		st.BytesRead += let.NBytes

		if let.NBytes <= 0 {
			st.ReadBuf = []byte{}
			nextAct, ioerr = st.Transporter.OnRead(st.Rudder, st.ReadBuf, nil)

		} else {
			nextAct, ioerr = st.Transporter.OnRead(st.Rudder, st.ReadBuf[0:let.NBytes], nil)
			//baylog.Debug("%s return read before len=%d", g, st.ReadPos)
			copy(st.ReadBuf, st.ReadBuf[let.NBytes:st.ReadPos])
			st.ReadPos = st.ReadPos - let.NBytes
			//baylog.Debug("%s return read len=%d", g, st.ReadPos)
		}

		break
	}

	if ioerr != nil {
		st.Transporter.OnError(st.Rudder, ioerr)
		nextAct = common.NEXT_SOCKET_ACTION_CLOSE
	}

	g.nextAction(st, nextAct, true)
}

func (g *GrandAgentImpl) onWrote(let *letter.WroteLetter) {
	st := let.State()

	baylog.Debug("%s wrote %d bytes rd=%s qlen=%d", g, let.NBytes, st.Rudder, len(st.WriteQueue))
	st.BytesWrote += let.NBytes

	if len(st.WriteQueue) == 0 {
		bayserver.FatalError(exception.NewSink("Write queue is empty"))
	}

	writeMore := true
	unit := st.WriteQueue[0]

	if len(unit.Buf) > 0 {
		baylog.Debug("Could not write enough data buf=%s", unit.Buf)
		writeMore = true

	} else {
		st.Multiplexer.ConsumeOldestUnit(st)
	}

	st.WriteLock.Lock()
	if len(st.WriteQueue) == 0 {
		writeMore = false
		st.Writing = false
	}
	st.WriteLock.Unlock()

	if writeMore {
		st.Multiplexer.NextWrite(st)

	} else {
		if st.Finale {
			// Close
			baylog.Debug("%s finale return Close", g)
			g.nextAction(st, common.NEXT_SOCKET_ACTION_CLOSE, false)

		} else {
			// Write off
			st.Multiplexer.CancelWrite(st)
		}
	}
}

func (g *GrandAgentImpl) onClosed(let *letter.ClosedLetter) {
	st := let.State()
	if st.Closed {
		baylog.Debug("%s Rudder is already closed: rd=%s", g, st.Rudder)
		return
	}

	g.netMultiplexer.RemoveRudderState(st.Rudder)

	// Clear write queue
	for st.Multiplexer.ConsumeOldestUnit(st) {

	}

	if st.Transporter != nil {
		st.Transporter.OnClosed(st.Rudder)
	}

	st.Closed = true
	st.Access()
}

func (g *GrandAgentImpl) onError(let *letter.ErrorLetter) exception.Exception {
	baylog.Debug("%s onError", g)

	st := let.State()
	var ex = let.Err

	if _, ok := ex.(exception2.HttpException); ok {
		if st.Transporter != nil {
			st.Transporter.OnError(st.Rudder, ex)
		} else {
			baylog.ErrorE(ex, "Error letter")
		}
		if _, ok := st.Rudder.(*impl.ListenerRudder); !ok {
			g.nextAction(st, common.NEXT_SOCKET_ACTION_CLOSE, true)
		}
		return nil

	} else if _, ok := ex.(exception.IOException); ok {
		if st.Transporter != nil {
			st.Transporter.OnError(st.Rudder, ex)
		} else {
			baylog.ErrorE(ex, "Error letter")
		}
		if _, ok := st.Rudder.(*impl.ListenerRudder); !ok {
			g.nextAction(st, common.NEXT_SOCKET_ACTION_CLOSE, true)
		}
		return nil

	} else {
		return ex
	}
}

func (g *GrandAgentImpl) nextAction(st *common.RudderState, act common.NextSocketAction, reading bool) {
	baylog.Debug("%s next action: %d (reading=%t)", g, act, reading)
	cancel := false

	switch act {
	case common.NEXT_SOCKET_ACTION_CONTINUE:
		if reading {
			st.Multiplexer.NextRead(st)
		}
		break

	case common.NEXT_SOCKET_ACTION_READ:
		st.Multiplexer.NextRead(st)
		break

	case common.NEXT_SOCKET_ACTION_WRITE:
		if reading {
			cancel = true
		}
		break

	case common.NEXT_SOCKET_ACTION_CLOSE:
		if reading {
			cancel = true
		}
		st.Multiplexer.ReqClose(st.Rudder)
		break

	case common.NEXT_SOCKET_ACTION_SUSPEND:
		if reading {
			cancel = true
		}
		break
	}

	if cancel {
		st.Multiplexer.CancelRead(st)
		st.ReadLock.Lock()
		baylog.Debug("%s Reading off %s", g, st.Rudder)
		st.Reading = false
		st.ReadLock.Unlock()
	}

	st.Access()
}

func (g *GrandAgentImpl) countPostpone() int {
	return len(g.postponeQueue)
}

func (g *GrandAgentImpl) CatchUp() {
	g.postponeQueueLock.Lock()
	baylog.Debug("%s catchUp", g)
	if len(g.postponeQueue) > 0 {
		pp := g.postponeQueue[0]
		g.postponeQueue = arrayutil.RemoveAt(g.postponeQueue, 0)
		pp.Run()
	}
	g.postponeQueueLock.Unlock()
}

/****************************************/
/* static functions                    */
/****************************************/

func Init() {
	agent.Get = _get
	agent.AddLifeCycleListener = _addLifecycleListener
	agent.Add = _add
	agent.Init = _init
}

func _init(agentIds []int, nShips int) {
	agentCount = len(agentIds)
	maxShips = nShips
}

func _add(
	agentId int, anchorable bool) agent.GrandAgent {

	if agentId == -1 {
		maxAgentId++
		agentId = maxAgentId
	}
	baylog.Debug("Add agent: id=%d", agentId)

	if agentId > maxAgentId {
		maxAgentId = agentId
	}

	agt := NewGrandAgent(agentId, maxShips, anchorable)
	agents[agentId] = agt

	for _, lis := range listeners {
		lis.Add(agentId)
	}

	return agt
}

func _get(agtId int) agent.GrandAgent {
	return agents[agtId]
}

func _addLifecycleListener(lis common.LifecycleListener) {
	listeners = append(listeners, lis)
}
