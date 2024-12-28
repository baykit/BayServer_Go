package cgi

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/cgiutil"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type CgiDockerSub interface {
	CreateProcess(env map[string]string) (*exec.Cmd, exception.IOException)
}

const DEFAULT_TIMEOUT_SEC = 60

type CgiDockerBase struct {
	*base.ClubBase
	cgiDockerSub CgiDockerSub
	interpreter  string
	scriptBase   string
	docRoot      string
	timeoutSec   int
	maxProcesses int

	processCount int
	waitCount    int
	lock         sync.Mutex
}

func NewCgiDockerBase(sub CgiDockerSub) *CgiDockerBase {
	b := &CgiDockerBase{
		interpreter: "",
		scriptBase:  "",
		docRoot:     "",
		timeoutSec:  DEFAULT_TIMEOUT_SEC,
	}
	b.ClubBase = base.NewClubBase(b)
	b.cgiDockerSub = sub

	var _ docker.Club = b // implement check

	return b
}

func (d *CgiDockerBase) String() string {
	return "CgiDocker"
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (d *CgiDockerBase) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {

	switch strings.ToLower(kv.Key) {
	case "interpreter":
		d.interpreter = kv.Value

	case "scriptbase":
		d.scriptBase = kv.Value

	case "docroot":
		d.docRoot = kv.Value

	case "timeout":
		var err error
		d.timeoutSec, err = strconv.Atoi(kv.Value)
		if err != nil {
			baylog.ErrorE(exception.NewExceptionFromError(err), "")
		}

	case "maxprocesses":
		var err error
		d.maxProcesses, err = strconv.Atoi(kv.Value)
		if err != nil {
			baylog.ErrorE(exception.NewExceptionFromError(err), "")
		}

	default:
		var cerr exception2.ConfigException
		_, cerr = d.ClubBase.InitKeyVal(kv)
		if cerr != nil {
			return false, cerr
		}
	}

	return true, nil
}

/****************************************/
/* Implements Club                      */
/****************************************/

func (d *CgiDockerBase) Arrive(tur tour.Tour) exception2.HttpException {
	if strings.Contains(tur.Req().Uri(), "..") {
		return exception2.NewHttpException(httpstatus.FORBIDDEN, tur.Req().Uri())
	}

	base := d.scriptBase
	town := tur.Town().(docker.Town)
	if base == "" {
		base = town.Location()
	}

	if base == "" {
		return exception2.NewHttpException(httpstatus.INTERNAL_SERVER_ERROR, "% scriptBase of cgi docker or location of town is not specified", town)
	}

	root := d.docRoot
	if root == "" {
		root = town.Location()
	}

	if root == "" {
		return exception2.NewHttpException(httpstatus.INTERNAL_SERVER_ERROR, "% docRoot of cgi docker or location of town is not specified", town)
	}

	env := cgiutil.GetEnv(town.Name(), root, base, tur)
	if bayserver.Harbor().TraceHeader() {
		for name, value := range env {
			baylog.Info("%s cgi: env: %s=%s", tur, name, value)
		}
	}

	fileName := env[cgiutil.SCRIPT_FILENAME]
	if !sysutil.IsFile(fileName) {
		return exception2.NewHttpException(httpstatus.NOT_FOUND, fileName)
	}

	handler := NewCgiReqContentHandler(d, tur, env)
	tur.Req().SetReqContentHandler(handler)
	handler.reqStartTour()

	return nil
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (d *CgiDockerBase) TimeoutSec() int {
	return d.timeoutSec
}

func (d *CgiDockerBase) GetWaitCount() int {
	return d.waitCount
}

func (d *CgiDockerBase) AddProcessCount() bool {
	d.lock.Lock()
	var ret bool
	if d.maxProcesses <= 0 || d.processCount < d.maxProcesses {
		d.processCount++
		baylog.Debug("%s Process count: %d", d, d.processCount)
		ret = true

	} else {
		d.waitCount++
		ret = false
	}
	d.lock.Unlock()
	return ret
}

func (d *CgiDockerBase) SubProcessCount() {
	d.lock.Lock()
	d.processCount--
	d.lock.Unlock()
}

func (d *CgiDockerBase) SubWaitCount() {
	d.lock.Lock()
	d.waitCount--
	d.lock.Unlock()
}
