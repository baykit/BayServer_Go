package multiplexer

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"errors"
	"net"
	"sync"
	"syscall"
	"time"
)

type JobMultiplexer struct {
	*MultiplexerBase
	agent          agent.GrandAgent
	anchorable     bool
	operationsLock sync.Mutex
	fdStateMapLock sync.Mutex
}

func NewJobMultiplexer(agent agent.GrandAgent, anchorable bool) common.Multiplexer {
	mpx := &JobMultiplexer{
		agent:          agent,
		anchorable:     anchorable,
		operationsLock: sync.Mutex{},
		fdStateMapLock: sync.Mutex{},
	}
	mpx.MultiplexerBase = NewMultiplexerBase(agent)

	agent.AddTimerHandler(mpx)

	// interface check
	var _ common.Multiplexer = mpx

	return mpx
}

func (mpx *JobMultiplexer) String() string {
	return mpx.agent.String()
}

/****************************************/
/* Implements Multiplexer               */
/****************************************/

func (mpx *JobMultiplexer) AddRudderState(rd rudder.Rudder, st *common.RudderState) {
	mpx.MultiplexerBase.AddRudderState(mpx, rd, st)
}

func (mpx *JobMultiplexer) ReqAccept(rd rudder.Rudder) {
	baylog.Debug("%s reqAccept rd=%s isShutdown=%v", mpx, rd, mpx.agent.Aborted)
	if mpx.agent.Aborted() {
		return
	}
	st := mpx.FindRudderState(rd)
	if st == nil {
		return
	}
	baylog.Debug("%s reqAccept rd in state=%s", mpx, st.Rudder)

	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		if mpx.agent.Aborted() {
			return
		}

		for {
			con, err := impl.GetListener(rd).Accept()
			if err != nil {
				if errors.Is(err, syscall.EAGAIN) {
					continue
				} else {
					ioerr := exception.NewIOExceptionFromError(err)
					mpx.agent.SendErrorLetter(st, ioerr, true)
					return
				}
			}

			p := bayserver.AnchorablePortMap()[rd]
			baylog.Debug("%s Accepted: server_rd=%s client_con=%s", mpx.agent, rd, con)
			if p.Secure() {

				go func() {
					defer func() {
						bayserver.BDefer()
					}()

					// Handshake in go routine
					sslCon, ioerr := p.GetSecureConn(con)
					if ioerr != nil {
						mpx.agent.SendErrorLetter(st, ioerr, true)
						_ = con.Close()
						return
					}
					mpx.handleAccept(st, sslCon)
				}()

			} else {
				mpx.handleAccept(st, con)
			}
		}
	}()
}

func (mpx *JobMultiplexer) ReqConnect(rd rudder.Rudder, adr net.Addr) exception.IOException {
	baylog.Debug("%s reqConnect rd=%s", mpx.agent, rd)

	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		st := mpx.FindRudderState(rd)
		if st == nil || st.Closing {
			// channel is already closed
			baylog.Debug("%s Channel is already closed: stm=%d", mpx.agent, rd)
			return
		}

		var ioerr exception.IOException = nil
		for { // try catch
			var tcpRd = rd.(*impl.TcpConnRudder)
			var tcpAddr = adr.(*net.TCPAddr)

			con, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				baylog.Debug("Connect error: rd=%s err=%s", tcpRd, err.Error())
				ioerr = exception.NewIOExceptionFromError(err)
				break
			}

			tcpRd.Conn = con
			baylog.Debug("Connected: rd=%s", tcpRd)
			mpx.agent.SendConnectedLetter(st, true)
			return
		}

		if ioerr != nil {
			mpx.agent.SendErrorLetter(st, ioerr, true)
		}

	}()

	st := mpx.FindRudderState(rd)
	if st == nil {
		return exception.NewIOException("State not found")
	}
	st.Access()
	return nil
}

func (mpx *JobMultiplexer) ReqRead(rd rudder.Rudder) {
	baylog.Debug("%s reqRead rd=%s", mpx.agent, rd)

	st := mpx.FindRudderState(rd)
	if st == nil {
		baylog.Debug("%s Unknown Rudder: %s", mpx.agent, rd)
		return

	} else if st.Closed {
		baylog.Debug("%s Rudder is already closed: rd=%s", mpx.agent, rd)
		return

	}

	st.Access()

	needRead := false
	st.ReadLock.Lock()
	if !st.Reading {
		st.Reading = true
		needRead = true
	}
	st.ReadLock.Unlock()
	baylog.Debug("%s needRead=%v", mpx.agent, needRead)

	if needRead {
		mpx.NextRead(st)
	}

}

func (mpx *JobMultiplexer) ReqWrite(
	rd rudder.Rudder,
	buf []byte,
	adr net.Addr,
	tag interface{},
	lis common.DataConsumeListener) exception.IOException {

	baylog.Debug("%s reqWrite rd=%s len=%d tag=%s", mpx.agent, rd, len(buf), tag)
	//baylog.Debug("%s data=%s", string(buf))

	st := mpx.FindRudderState(rd)
	if st == nil {
		return exception.NewIOException("%s Unknown Rudder: %s", mpx.agent, rd)

	} else if st.Closed {
		return exception.NewIOException("%s Rudder is already closed: rd=%s", mpx.agent, rd)

	}

	unit := common.NewWriteUnit(buf, adr, tag, lis)
	st.QueueLock.Lock()
	st.WriteQueue = append(st.WriteQueue, unit)
	st.QueueLock.Unlock()

	needWrite := false
	st.WriteLock.Lock()
	if !st.Writing {
		st.Writing = true
		needWrite = true
	}
	st.WriteLock.Unlock()

	if needWrite {
		mpx.NextWrite(st)
	}

	st.Access()
	return nil
}

func (mpx *JobMultiplexer) ReqEnd(rd rudder.Rudder) {
	st := mpx.FindRudderState(rd)
	if st == nil {
		return
	}

	st.End()
	st.Access()
}

func (mpx *JobMultiplexer) ReqClose(rd rudder.Rudder) {
	st := mpx.FindRudderState(rd)
	baylog.Debug("%s reqClose rd=%s state=%s", mpx.agent, rd, st)
	if st == nil {
		baylog.Warn("%s Unknown Rudder: %s", mpx.agent, rd)
		return

	} else if st.Closed {
		baylog.Warn("%s Rudder is closed: %s", mpx.agent, rd)
		return

	}

	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		st = mpx.FindRudderState(rd)
		if st == nil {
			baylog.Warn("%s Unknown Rudder: %s", mpx.agent, rd)
			return

		} else if st.Closed {
			baylog.Warn("%s Rudder is closed: %s", mpx.agent, rd)
			return

		}

		mpx.CloseRudder(st)

		mpx.agent.SendClosedLetter(st, true)
	}()

	st.Access()
}

func (mpx *JobMultiplexer) CancelRead(st *common.RudderState) {
}

func (mpx *JobMultiplexer) CancelWrite(st *common.RudderState) {
}

func (mpx *JobMultiplexer) NextAccept(st *common.RudderState) {
	mpx.ReqAccept(st.Rudder)
}

func (mpx *JobMultiplexer) NextRead(st *common.RudderState) {
	baylog.Debug("%s nextRead rd=%s", mpx.agent, st.Rudder)

	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		var ioerr exception.IOException = nil

		if st == nil || st.Closing {
			// channel is already closed
			baylog.Debug("%s Channel is already closed: rd=%s", mpx.agent, st.Rudder)
			return
		}

		var n int

		baylog.Debug("%s Try to Read (rd=%s) (readPos=%d bufLen=%d buf=%s)", mpx.agent, st.Rudder, st.ReadPos, len(st.ReadBuf), st.ReadBuf[0:st.ReadPos])
		n, ioerr = st.Rudder.Read(st.ReadBuf)
		//baylog.Debug("%s Read %d bytes (rd=%s) (buf=%s)", mpx.agent, n, st.Rudder, st.ReadBuf[0:n])

		if ioerr != nil {
			if st.Closed {
				baylog.DebugE(ioerr, "%s Closed by another thread: %s", mpx, st.Rudder)
			} else {
				mpx.agent.SendErrorLetter(st, ioerr, true)
			}

		} else {
			st.ReadPos += n
			mpx.agent.SendReadLetter(st, n, "", true)
		}

	}()
}

func (mpx *JobMultiplexer) NextWrite(st *common.RudderState) {
	baylog.Debug("%s nextWrite rd=%s", mpx.agent, st.Rudder)

	go func() {
		defer func() {
			bayserver.BDefer()
		}()

		if st == nil || st.Closing || st.Closed {
			// channel is already closed
			baylog.Debug("%s Channel is already closed: rd=%s", mpx.agent, st.Rudder)
			return
		}

		if len(st.WriteQueue) == 0 {
			baylog.Fatal("%s Write queue is empty rd=%s", mpx.agent, st.Rudder)
			return
		}

		u := st.WriteQueue[0]
		baylog.Debug("%s Try to write: rd=%s pkt=%s buflen=%d closed=%v", mpx, st.Rudder, u.Tag, len(u.Buf), st.Closed)
		err := false

		//baylog.Debug("%s rd=%s data: %s", string(unit.Buf))
		var n = 0
		if !st.Closed && len(u.Buf) > 0 {
			var ioerr exception.IOException
			n, ioerr = st.Rudder.Write(u.Buf)
			//baylog.Trace("%s rd=%s wrote %d bytes", mpx, st.Rudder, n)
			if ioerr != nil {
				mpx.agent.SendErrorLetter(st, ioerr, true)
				err = true
			} else if n != len(u.Buf) {
				mpx.agent.SendErrorLetter(st, exception.NewIOException("Could not write enough data"), true)
				err = true
			} else {
				u.Buf = u.Buf[0:0] // make buffer empty
			}
		}

		if !err {
			mpx.agent.SendWroteLetter(st, n, true)
		}

	}()
}

func (mpx *JobMultiplexer) Shutdown() {
	mpx.closeAll()
}

func (mpx *JobMultiplexer) IsNonBlocking() bool {
	return false
}

func (mpx *JobMultiplexer) OnBusy() {

}

func (mpx *JobMultiplexer) OnFree() {
	if mpx.agent.Aborted() {
		return
	}

	for rd := range bayserver.AnchorablePortMap() {
		mpx.ReqAccept(rd)
	}
}

func (mpx *JobMultiplexer) CloseTimeoutSockets() {
	if len(mpx.rudders) == 0 {
		return
	}

	closeList := []*common.RudderState{}
	mpx.lock.Lock()
	now := time.Now().Unix()
	for _, st := range mpx.rudders {
		if st.Transporter != nil && st.Transporter.CheckTimeout(st.Rudder, int(now-st.LastAccessTime)) {
			baylog.Debug("%s found timed out socket: rd=%s", mpx.agent, st.Rudder)
			closeList = append(closeList, st)
		}
	}

	mpx.lock.Unlock()

	for _, st := range closeList {
		mpx.ReqClose(st.Rudder)
	}
}

/****************************************/
/* Implements TimerHandler              */
/****************************************/

func (mpx *JobMultiplexer) OnTimer() {
	mpx.CloseTimeoutSockets()
}

/****************************************/
/* Public functions                     */
/****************************************/

/****************************************/
/* Private functions                    */
/****************************************/

func (mpx *JobMultiplexer) handleAccept(st *common.RudderState, con net.Conn) {
	rd := impl.NewTcpConnRudder(con)
	baylog.Debug("%s Accepted: server rd=%s client rd=%s", mpx.agent, st.Rudder, rd)

	if mpx.agent.Aborted() {
		baylog.Error("%s Agent is not alive (close)", mpx.agent)
		_ = con.Close()

	} else {
		mpx.agent.SendAcceptedLetter(st, rd, true)
	}
}
