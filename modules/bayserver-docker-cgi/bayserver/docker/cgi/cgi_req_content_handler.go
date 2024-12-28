package cgi

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/agent/multiplexer"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/inboundship"
	"bayserver-core/baykit/bayserver/docker"
	impl2 "bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/tour/impl"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"errors"
	"io"
	"os/exec"
	"syscall"
	"time"
)

type CgiReqContentHandler struct {
	cgiDocker    *CgiDockerBase
	tour         tour.Tour
	tourId       int
	available    bool
	cmd          *exec.Cmd
	inPipe       io.WriteCloser
	outPipe      io.ReadCloser
	errPipe      io.ReadCloser
	stdOutClosed bool
	stdErrClosed bool
	lastAccess   int64
	env          map[string]string
}

func NewCgiReqContentHandler(dkr *CgiDockerBase, tur tour.Tour, env map[string]string) *CgiReqContentHandler {
	h := CgiReqContentHandler{
		cgiDocker: dkr,
		tour:      tur,
		tourId:    tur.TourId(),
		env:       env,
	}
	var _ tour.ReqContentHandler = &h // Cast check
	h.Access()
	return &h
}

/****************************************/
/* Implements ReqContentHandler         */
/****************************************/

func (c *CgiReqContentHandler) OnReadReqContent(tur tour.Tour, buf []byte, start int, length int, lis tour.ContentConsumeListener) exception2.IOException {
	baylog.Debug("%s CGI:onReadReqContent: len=%d", tur, length)
	_, err := c.inPipe.Write(buf[start : start+length])
	if err != nil {
		return exception2.NewIOExceptionFromError(err)
	}
	tur.Req().Consumed(impl.TOUR_ID_NOCHECK, length, lis)
	c.Access()
	return nil
}

func (c *CgiReqContentHandler) OnEndReqContent(tur tour.Tour) (exception2.IOException, exception.HttpException) {
	baylog.Debug("%s CGI:endReqContent", tur)
	c.Access()
	return nil, nil
}

func (c *CgiReqContentHandler) OnAbortReq(tur tour.Tour) bool {
	baylog.Debug("%s CGI:abortReq", tur)
	if c.cmd != nil {
		err := c.inPipe.Close()
		if err != nil {
			baylog.ErrorE(exception2.NewIOExceptionFromError(err), "")
		}
	}
	return false // not aborted immediately
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (c *CgiReqContentHandler) Run() {
	c.cgiDocker.SubWaitCount()
	baylog.Info("%s challenge postponed tour wait count=%d", c.tour, c.cgiDocker.GetWaitCount())
	c.reqStartTour()
}

func (c *CgiReqContentHandler) StartTour() {
	c.available = false

	var ioerr exception2.IOException = nil
	for { // try catch
		c.cmd, ioerr = c.cgiDocker.cgiDockerSub.CreateProcess(c.env)
		if ioerr != nil {
			break
		}

		var err error
		c.inPipe, err = c.cmd.StdinPipe()
		if err != nil {
			break
		}
		c.outPipe, err = c.cmd.StdoutPipe()
		if err != nil {
			break
		}
		c.errPipe, err = c.cmd.StderrPipe()
		if err != nil {
			break
		}

		err = c.cmd.Start()
		if err != nil {
			break
		}

		break
	}

	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
		ioerr = c.tour.Res().SendError(c.tourId, httpstatus.INTERNAL_SERVER_ERROR, "Process Error", ioerr)
		if ioerr != nil {
			baylog.ErrorE(ioerr, "")
		}
	}

	c.stdOutClosed = false
	c.stdErrClosed = false

	outRd := impl2.NewReadCloserRudder(c.outPipe)
	errRd := impl2.NewReadCloserRudder(c.errPipe)

	sip := c.tour.Ship().(ship.Ship)
	agt := agent.Get(sip.AgentId())
	var mpx common.Multiplexer = nil

	mpxType := bayserver.Harbor().CgiMultiplexer()
	switch mpxType {
	case docker.MULTI_PLEXER_TYPE_JOB:
		mpx = agt.JobMultiplexer()

	default:
		bayserver.FatalError(exception2.NewSink("Unsupported multiplexer type: %s", docker.GetMultiplexerTypeName(mpxType)))

	}

	outShip := NewCgiStdOutShip()
	bufsize := c.tour.Ship().(inboundship.InboundShip).ProtocolHandler().MaxResPacketDataSize()
	//baylog.Debug("bufsize=%d", bufsize)

	outTp := multiplexer.NewPlainTransporter(
		agt.NetMultiplexer(),
		outShip,
		false,
		bufsize,
		false)
	outShip.Init(outRd, sip.AgentId(), c.tour, outTp, c)

	mpx.AddRudderState(
		outRd,
		common.NewRudderState(
			outRd,
			outTp))

	sipId := outShip.ShipId()
	c.tour.Res().SetConsumeListener(func(len int, resume bool) {
		if resume {
			outShip.ResumeRead(sipId)
		}
	})

	agt.JobMultiplexer().ReqRead(outRd)

	errShip := NewCgiStdErrShip()
	errShip.Init(errRd, sip.AgentId(), c)
	errTp := multiplexer.NewPlainTransporter(
		agt.NetMultiplexer(),
		errShip,
		false,
		bufsize,
		false)

	mpx.AddRudderState(
		errRd,
		common.NewRudderState(
			errRd,
			errTp))

	agt.JobMultiplexer().ReqRead(errRd)

}

func (c *CgiReqContentHandler) StdOutClosed() {
	c.stdOutClosed = true
	if c.stdOutClosed && c.stdErrClosed {
		c.processFinished()
	}
}

func (c *CgiReqContentHandler) StdErrClosed() {
	c.stdErrClosed = true
	if c.stdOutClosed && c.stdErrClosed {
		c.processFinished()
	}
}

func (c *CgiReqContentHandler) Access() {
	c.lastAccess = time.Now().Unix()
}

func (c *CgiReqContentHandler) Timeout() bool {
	if c.cgiDocker.TimeoutSec() <= 0 {
		return false
	}

	baylog.Debug("%s Check CGI timeout: now=%d, last=%d", c.tour, time.Now().Unix(), c.lastAccess)

	durationSec := int(time.Now().Unix() - c.lastAccess)
	baylog.Debug("%s Check CGI timeout: dur=%d, timeout=%d", c.tour, durationSec, c.cgiDocker.TimeoutSec())
	return durationSec > c.cgiDocker.TimeoutSec()
}

/****************************************/
/* Prvate functions                     */
/****************************************/

func (c *CgiReqContentHandler) reqStartTour() {
	if c.cgiDocker.AddProcessCount() {
		baylog.Info("%s start tour: wait count=%d", c.tour, c.cgiDocker.GetWaitCount())
		c.StartTour()

	} else {
		baylog.Warn("%s Cannot start tour: wait count=%d", c.tour, c.cgiDocker.GetWaitCount())
		agt := agent.Get(c.tour.Ship().(ship.Ship).AgentId())
		agt.AddPostpone(c)
	}
	c.Access()
}

func (c *CgiReqContentHandler) processFinished() {
	baylog.Debug("%s process_finished()", c.tour)
	agtId := c.tour.Ship().(ship.Ship).AgentId()

	hasErr := false
	err := c.cmd.Wait()
	if err != nil {
		hasErr = true
		var exiterr *exec.ExitError
		if errors.As(err, &exiterr) {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				baylog.Error("CGI process error: exit code=%d", status.ExitStatus())
			} else {
				baylog.ErrorE(exception2.NewIOExceptionFromError(err), "")
			}
		}
	} else {
		baylog.Debug("CGI Process finished (^o^)")
	}

	var ioerr exception2.IOException = nil
	if hasErr {
		ioerr = c.tour.Res().SendError(c.tourId, httpstatus.INTERNAL_SERVER_ERROR, "Exec failed", nil)

	} else {
		ioerr = c.tour.Res().EndResContent(c.tourId)
	}

	if ioerr != nil {
		baylog.ErrorE(ioerr, "")
	}

	c.cgiDocker.SubProcessCount()
	if c.cgiDocker.GetWaitCount() > 0 {
		baylog.Warn("agt#%d Catch up postponed process: process wait count=%d", agtId, c.cgiDocker.GetWaitCount())
		agt := agent.Get(agtId)
		agt.ReqCatchUp()
	}
}
