package builtin

import (
	"bayserver-core/baykit/bayserver/agent"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/docker/builtin/logitems"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/ship"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"os"
	"path"
	"strconv"
	"strings"
)

/****************************************/
/* type LogAgentListener                */
/****************************************/

type LogAgentListener struct {
	docker *BuiltInLogDocker
}

func NewLogAgentListener(dkr *BuiltInLogDocker) common.LifecycleListener {
	return &LogAgentListener{
		docker: dkr,
	}
}

func (lis *LogAgentListener) Add(agentId int) {
	fileName := lis.docker.filePrefix + "_" + strconv.Itoa(agentId) + "." + lis.docker.fileExt
	var ioerr exception.IOException = nil
	for { // try-catch
		agt := agent.Get(agentId)
		var rd rudder.Rudder
		var mpx common.Multiplexer
		switch bayserver.Harbor().LogMultiplexer() {
		case docker.MULTI_PLEXER_TYPE_TAXI:
			{

			}
		case docker.MULTI_PLEXER_TYPE_JOB:
			{
				file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE, 0755)
				if err != nil {
					ioerr = exception.NewIOExceptionFromError(err)
					break
				}
				rd = impl.NewFileRudder(file)
				mpx = agt.JobMultiplexer()
			}

		default:
			bayserver.FatalError(exception.NewSink("Multiplexer type not supported: %s", bayserver.Harbor().LogMultiplexer()))
		}

		st := common.NewRudderState(rd, nil)
		st.BytesWrote = -1
		mpx.AddRudderState(rd, st)
		lis.docker.multiplexers = append(lis.docker.multiplexers, mpx)
		lis.docker.rudders = append(lis.docker.rudders, rd)
		break
	}

	if ioerr != nil {
		baylog.Fatal(baymessage.Get(symbol.INT_CANNOT_OPEN_LOG_FILE, fileName))
		baylog.FatalE(ioerr, "")
	}
}

func (lis *LogAgentListener) Remove(agentId int) {
	rd := lis.docker.rudders[agentId-1]
	lis.docker.multiplexers[agentId-1].ReqClose(rd)
	lis.docker.multiplexers[agentId-1] = nil
	lis.docker.rudders[agentId-1] = nil
}

/****************************************/
/* type BuildInLogDocker                */
/****************************************/

var logItemFactoryMap = map[string]LogItemFactory{
	"a":  func() logitems.LogItem { return logitems.NewRemoteIpItem() },
	"A":  func() logitems.LogItem { return logitems.NewServerIpItem() },
	"b":  func() logitems.LogItem { return logitems.NewRequestBytesItem2() },
	"B":  func() logitems.LogItem { return logitems.NewRequestBytesItem1() },
	"c":  func() logitems.LogItem { return logitems.NewConnectionStatusItem() },
	"e":  func() logitems.LogItem { return logitems.NewNullItem() },
	"h":  func() logitems.LogItem { return logitems.NewRemoteHostItem() },
	"i":  func() logitems.LogItem { return logitems.NewRequestHeaderItem() },
	"l":  func() logitems.LogItem { return logitems.NewRemoteLogItem() },
	"m":  func() logitems.LogItem { return logitems.NewMethodItem() },
	"n":  func() logitems.LogItem { return logitems.NewNullItem() },
	"o":  func() logitems.LogItem { return logitems.NewResponseHeaderItem() },
	"p":  func() logitems.LogItem { return logitems.NewPortItem() },
	"P":  func() logitems.LogItem { return logitems.NewNullItem() },
	"q":  func() logitems.LogItem { return logitems.NewQueryStringItem() },
	"r":  func() logitems.LogItem { return logitems.NewStartLineItem() },
	"s":  func() logitems.LogItem { return logitems.NewStatusItem() },
	">s": func() logitems.LogItem { return logitems.NewStatusItem() },
	"t":  func() logitems.LogItem { return logitems.NewTimeItem() },
	"T":  func() logitems.LogItem { return logitems.NewIntervalItem() },
	"u":  func() logitems.LogItem { return logitems.NewRemoteUserItem() },
	"U":  func() logitems.LogItem { return logitems.NewRequestUrlItem() },
	"v":  func() logitems.LogItem { return logitems.NewServerNameItem() },
	"V":  func() logitems.LogItem { return logitems.NewNullItem() },
}

type BuiltInLogDocker struct {
	*base.DockerBase

	/** Log file name parts */
	filePrefix string
	fileExt    string

	/** Log format */
	format string

	/** Log items */
	logItems []logitems.LogItem

	rudders []rudder.Rudder

	/** Multiplexer to write to file */
	multiplexers []common.Multiplexer
}

func NewBuiltInLogDocker() docker.Log {
	dkr := &BuiltInLogDocker{}
	dkr.DockerBase = base.NewDockerBase(dkr)
	dkr.logItems = make([]logitems.LogItem, 0)
	dkr.rudders = make([]rudder.Rudder, 0)
	dkr.multiplexers = make([]common.Multiplexer, 0)
	return dkr
}

func (d *BuiltInLogDocker) String() string {
	return "BuiltInLogDocker(" + d.filePrefix + "." + d.fileExt + ")"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *BuiltInLogDocker) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	err := d.DockerBase.Init(elm, parent)
	if err != nil {
		return err
	}

	p := strings.Index(elm.Arg, ".")
	if p == -1 {
		d.filePrefix = elm.Arg
		d.fileExt = ""
	} else {
		d.filePrefix = elm.Arg[:p]
		d.fileExt = elm.Arg[p+1:]
	}

	if d.format == "" {
		return exception2.NewConfigException(
			elm.FileName,
			elm.LineNo,
			baymessage.Get(symbol.CFG_INVALID_LOG_FORMAT, ""))
	}

	if !path.IsAbs(d.filePrefix) {
		d.filePrefix = path.Join(bayserver.BservHome(), d.filePrefix)
	}

	logDir := path.Dir(d.filePrefix)
	if !sysutil.IsDirectory(logDir) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			baylog.ErrorE(exception.NewExceptionFromError(err), "")
			return exception2.NewConfigException(
				elm.FileName,
				elm.LineNo,
				"Cannot create directory")
		}
	}

	// Parse format
	d.logItems, err = d.compile(d.format, []logitems.LogItem{}, elm.FileName, elm.LineNo)
	if err != nil {
		return err
	}

	agent.AddLifeCycleListener(NewLogAgentListener(d))

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (d *BuiltInLogDocker) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	return true, nil
}

func (d *BuiltInLogDocker) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {
	switch strings.ToLower(kv.Key) {
	default:
		return d.DockerBase.DefaultInitKeyVal(kv)

	case "format":
		d.format = kv.Value
	}
	return true, nil
}

/****************************************/
/* Implements Log                       */
/****************************************/

func (d *BuiltInLogDocker) Log(tur tour.Tour) exception.IOException {

	line := ""
	for _, logItem := range d.logItems {
		item := logItem.GetItem(tur)
		if item == "" {
			line += "-"
		} else {
			line += item
		}
	}

	// If threre are message to write, write it
	if len(line) > 0 {
		bytes := []byte(line)
		return d.multiplexers[tur.Ship().(ship.Ship).AgentId()-1].ReqWrite(
			d.rudders[tur.Ship().(ship.Ship).AgentId()-1], bytes, nil, "log", nil)
	}

	return nil
}

/****************************************/
/* Private functions                    */
/****************************************/

/**
 * Compile format pattern
 */
func (d *BuiltInLogDocker) compile(str string, items []logitems.LogItem, fileName string, lineNo int) ([]logitems.LogItem, exception2.ConfigException) {

	// Find control code
	pos := strings.Index(str, "%")

	if pos != -1 {
		text := str[0:pos]
		items = append(items, logitems.NewTextItem(text))
		var err exception2.ConfigException
		items, err = d.compileCtl(str[pos+1:], items, fileName, lineNo)
		if err != nil {
			return nil, err
		}

	} else {
		items = append(items, logitems.NewTextItem(str))
	}

	return items, nil
}

/**
 * Compile format pattern(Control code)
 */
func (d *BuiltInLogDocker) compileCtl(str string, items []logitems.LogItem, fileName string, lineNo int) ([]logitems.LogItem, exception2.ConfigException) {

	param := ""

	// if exists param
	if str[0] == '{' {
		// find close bracket
		pos := strings.Index(str, "}")
		if pos == -1 {
			return nil, exception2.NewConfigException(fileName, lineNo, baymessage.Get(symbol.CFG_INVALID_LOG_FORMAT, d.format))
		}
		param = str[1:pos]
		str = str[pos+1:]
	}

	ctlChar := ""
	hasError := false

	if len(str) == 0 {
		hasError = true
	}

	if !hasError {
		// get control char
		ctlChar = str[0:1]
		str = str[1:]

		if ctlChar == ">" {
			if len(str) == 0 {
				hasError = true

			} else {
				ctlChar = str[0:1]
				str = str[1:]
			}
		}
	}

	var fct LogItemFactory = nil
	if !hasError {
		fct = logItemFactoryMap[ctlChar]
		if fct == nil {
			hasError = true
		}
	}

	if hasError {
		return nil, exception2.NewConfigException(
			fileName,
			lineNo,
			baymessage.Get(symbol.CFG_INVALID_LOG_FORMAT, d.format+" (unknown control code: '%"+ctlChar+"')"))
	}

	item := fct()
	item.Init(param)
	items = append(items, item)
	return d.compile(str, items, fileName, lineNo)
}
