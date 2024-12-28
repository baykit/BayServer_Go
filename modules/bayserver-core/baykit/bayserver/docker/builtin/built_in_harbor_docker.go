package builtin

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/groups"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/strutil"
	"strings"
)

const DEFAULT_MAX_SHIPS = 256

const DEFAULT_GRAND_AGENTS = 0

const DEFAULT_TRAIN_RUNNERS = 8

const DEFAULT_TAXI_RUNNERS = 8

const DEFAULT_SOCKET_TIMEOUT_SEC = 300

const DEFAULT_KEEP_TIMEOUT_SEC = 20

const DEFAULT_TOUR_BUFFER_SIZE = 1024 * 1024 // 1M

const DEFAULT_CHARSET = "UTF-8"

const DEFAULT_CONTROL_PORT = -1

const DEFAULT_NET_MULTIPLEXER = docker.MULTI_PLEXER_TYPE_JOB

const DEFAULT_FILE_MULTIPLEXER = docker.MULTI_PLEXER_TYPE_JOB

const DEFAULT_LOG_MULTIPLEXER = docker.MULTI_PLEXER_TYPE_JOB

const DEFAULT_CGI_MULTIPLEXER = docker.MULTI_PLEXER_TYPE_JOB

const DEFAULT_RECIPIENT = docker.RECIPIENT_TYPE_PIPE

const DEFAULT_MULTI_CORE = true

const DEFAULT_GZIP_COMP = false

const DEFAULT_PID_FILE = "bayserver.pid"

type BuiltInHarborDocker struct {
	*base.DockerBase

	charset          string
	locale           *util.Locale
	grandAgents      int
	trainRunners     int
	taxiRunners      int
	maxShips         int
	socketTimeoutSec int
	keepTimeoutSec   int
	tourBufferSize   int
	traceHeader      bool
	trouble          docker.Trouble
	redirectFile     string
	gzipComp         bool
	controlPort      int
	multiCore        bool
	netMultiplexer   int
	fileMultiplexer  int
	logMultiplexer   int
	cgiMultiplexer   int
	recipient        int
	pidFile          string
}

func NewBuiltInHarborDocker() docker.Harbor {
	h := &BuiltInHarborDocker{}
	h.DockerBase = base.NewDockerBase(h)
	h.charset = DEFAULT_CHARSET
	h.locale = util.DefaultLocale()
	h.grandAgents = DEFAULT_GRAND_AGENTS
	h.trainRunners = DEFAULT_TRAIN_RUNNERS
	h.taxiRunners = DEFAULT_TAXI_RUNNERS
	h.maxShips = DEFAULT_MAX_SHIPS
	h.socketTimeoutSec = DEFAULT_SOCKET_TIMEOUT_SEC
	h.keepTimeoutSec = DEFAULT_KEEP_TIMEOUT_SEC
	h.tourBufferSize = DEFAULT_TOUR_BUFFER_SIZE
	h.traceHeader = false
	h.trouble = nil
	h.redirectFile = ""
	h.gzipComp = DEFAULT_GZIP_COMP
	h.controlPort = DEFAULT_CONTROL_PORT
	h.multiCore = DEFAULT_MULTI_CORE
	h.netMultiplexer = DEFAULT_NET_MULTIPLEXER
	h.fileMultiplexer = DEFAULT_FILE_MULTIPLEXER
	h.logMultiplexer = DEFAULT_LOG_MULTIPLEXER
	h.cgiMultiplexer = DEFAULT_CGI_MULTIPLEXER
	h.recipient = DEFAULT_RECIPIENT
	h.pidFile = DEFAULT_PID_FILE

	return h
}

func (d *BuiltInHarborDocker) String() string {
	return "BuiltInHarborDocker"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *BuiltInHarborDocker) Init(elm *bcf.BcfElement, parent docker.Docker) exception.ConfigException {
	err := d.DockerBase.Init(elm, parent)
	if err != nil {
		return err
	}

	if d.grandAgents <= 0 {
		d.grandAgents = 1
	}

	if d.trainRunners <= 0 {
		d.trainRunners = 1
	}

	if d.maxShips <= 0 {
		d.maxShips = DEFAULT_MAX_SHIPS
	}

	if d.maxShips < DEFAULT_MAX_SHIPS {
		d.maxShips = DEFAULT_MAX_SHIPS
		baylog.Warn(baymessage.Get(symbol.CFG_MAX_SHIPS_IS_TO_SMALL, d.maxShips))
	}

	if d.netMultiplexer == docker.MULTI_PLEXER_TYPE_TAXI ||
		d.netMultiplexer == docker.MULTI_PLEXER_TYPE_TRAIN ||
		d.netMultiplexer == docker.MULTI_PLEXER_TYPE_SPIN {

		baylog.Warn(
			baymessage.Get(
				symbol.CFG_NET_MULTIPLEXER_NOT_SUPPORTED,
				docker.GetMultiplexerTypeName(d.netMultiplexer),
				docker.GetMultiplexerTypeName(DEFAULT_NET_MULTIPLEXER)))
		d.netMultiplexer = DEFAULT_NET_MULTIPLEXER
	}

	if d.fileMultiplexer != docker.MULTI_PLEXER_TYPE_JOB {
		baylog.Warn(
			baymessage.Get(
				symbol.CFG_FILE_MULTIPLEXER_NOT_SUPPORTED,
				docker.GetMultiplexerTypeName(d.FileMultiplexer()),
				docker.GetMultiplexerTypeName(DEFAULT_FILE_MULTIPLEXER)))
		d.fileMultiplexer = DEFAULT_FILE_MULTIPLEXER
	}

	if d.cgiMultiplexer == docker.MULTI_PLEXER_TYPE_SPIDER ||
		d.cgiMultiplexer == docker.MULTI_PLEXER_TYPE_SPIN ||
		d.cgiMultiplexer == docker.MULTI_PLEXER_TYPE_PIGEON ||
		d.cgiMultiplexer == docker.MULTI_PLEXER_TYPE_TRAIN {
		baylog.Warn(exception.CreatePositionMessage(
			baymessage.Get(
				symbol.CFG_CGI_MULTIPLEXER_NOT_SUPPORTED,
				docker.GetMultiplexerTypeName(d.cgiMultiplexer),
				docker.GetMultiplexerTypeName(DEFAULT_CGI_MULTIPLEXER)),
			elm.FileName,
			elm.LineNo))
		d.cgiMultiplexer = DEFAULT_CGI_MULTIPLEXER
	}

	if d.netMultiplexer == docker.MULTI_PLEXER_TYPE_PIGEON || d.netMultiplexer == docker.MULTI_PLEXER_TYPE_JOB {
		baylog.Warn("Pigeon needs only one grand agent")
		d.grandAgents = 1
	}

	if d.netMultiplexer == docker.MULTI_PLEXER_TYPE_SPIDER &&
		d.recipient != docker.RECIPIENT_TYPE_SPIDER {
		baylog.Warn(exception.CreatePositionMessage(
			baymessage.Get(
				symbol.CFG_NET_MULTIPLEXER_DOES_NOT_SUPPORT_THIS_RECIPIENT,
				docker.GetMultiplexerTypeName(d.netMultiplexer),
				docker.GetRecipientTypeName(d.recipient),
				docker.GetRecipientTypeName(docker.RECIPIENT_TYPE_SPIDER)),
			elm.FileName,
			elm.LineNo))
		d.recipient = docker.RECIPIENT_TYPE_SPIDER
	}

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (d *BuiltInHarborDocker) InitDocker(dkr docker.Docker) (bool, exception.ConfigException) {
	if t, ok := dkr.(docker.Trouble); ok {
		d.trouble = t
		return true, nil
	} else {
		return d.DefaultInitDocker()
	}
}

func (d *BuiltInHarborDocker) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception.ConfigException) {
	var err exception2.Exception = nil
	switch strings.ToLower(kv.Key) {
	default:
		return d.DefaultInitKeyVal(kv)

	case "loglevel":
		baylog.SetLogLevel(kv.Value)

	case "charset":
		charset := strutil.ParseCharset(kv.Value)
		if charset != "" {
			d.charset = charset
		}

	case "locale":
		d.locale = util.ParseLocale(kv.Value)

	case "groups":
		{
			fname, err := bayserver.ParsePath(kv.Value)
			if err != nil {
				baylog.ErrorE(err, "")
				return false, nil
			}
			perr := groups.GroupsInit(fname)
			if perr != nil {
				return false, perr
			}
		}

	case "grandagents":
		d.grandAgents, err = strutil.ParseInt(kv.Value)

	case "trains":
		d.trainRunners, err = strutil.ParseInt(kv.Value)

	case "taxis", "taxies":
		d.taxiRunners, err = strutil.ParseInt(kv.Value)

	case "maxships":
		d.maxShips, err = strutil.ParseInt(kv.Value)

	case "timeout":
		d.socketTimeoutSec, err = strutil.ParseInt(kv.Value)

	case "keeptimeout":
		d.keepTimeoutSec, err = strutil.ParseInt(kv.Value)

	case "tourbuffersize":
		d.tourBufferSize, err = strutil.ParseSize(kv.Value)

	case "traceheader":
		d.traceHeader, err = strutil.ParseBool(kv.Value)

	case "redirectfile":
		d.redirectFile = kv.Value

	case "controlport":
		d.controlPort, err = strutil.ParseInt(kv.Value)

	case "multicore":
		d.multiCore, err = strutil.ParseBool(kv.Value)

	case "gzipcomp":
		d.gzipComp, err = strutil.ParseBool(kv.Value)

	case "netmultiplexer":
		d.netMultiplexer, err = docker.GetMultiPlexerType(kv.Value)
		if err != nil {
			baylog.ErrorE(err, "")
			return false, exception.NewConfigException(
				kv.FileName,
				kv.LineNo,
				baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
		}

	case "filemultiplexer":
		d.fileMultiplexer, err = docker.GetMultiPlexerType(kv.Value)
		if err != nil {
			baylog.ErrorE(err, "")
			return false, exception.NewConfigException(
				kv.FileName,
				kv.LineNo,
				baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
		}

	case "logmultiplexer":
		d.logMultiplexer, err = docker.GetMultiPlexerType(kv.Value)
		if err != nil {
			baylog.ErrorE(err, "")
			return false, exception.NewConfigException(
				kv.FileName,
				kv.LineNo,
				baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
		}

	case "cgimultiplexer":
		d.cgiMultiplexer, err = docker.GetMultiPlexerType(kv.Value)
		if err != nil {
			baylog.ErrorE(err, "")
			return false, exception.NewConfigException(
				kv.FileName,
				kv.LineNo,
				baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
		}

	case "recipient":
		d.recipient, err = docker.GetRecipientType(kv.Value)
		if err != nil {
			baylog.ErrorE(err, "")
			return false, exception.NewConfigException(
				kv.FileName,
				kv.LineNo,
				baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
		}

	case "pidfile":
		d.pidFile = kv.Value
	}

	if err != nil {
		baylog.ErrorE(err, "")
		return false, exception.NewConfigException(
			kv.FileName,
			kv.LineNo,
			baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE),
			kv.Value)
	}
	return true, nil
}

/****************************************/
/* Implements Harbor                    */
/****************************************/

func (d *BuiltInHarborDocker) Charset() string {
	return d.charset
}

func (d *BuiltInHarborDocker) Locale() *util.Locale {
	return d.locale
}

func (d *BuiltInHarborDocker) GrandAgents() int {
	return d.grandAgents
}

func (d *BuiltInHarborDocker) TrainRunners() int {
	return d.trainRunners
}

func (d *BuiltInHarborDocker) TaxiRunners() int {
	return d.taxiRunners
}

func (d *BuiltInHarborDocker) MaxShips() int {
	return d.maxShips
}

func (d *BuiltInHarborDocker) Trouble() docker.Trouble {
	return d.trouble
}

func (d *BuiltInHarborDocker) SocketTimeoutSec() int {
	return d.socketTimeoutSec
}

func (d *BuiltInHarborDocker) KeepTimeoutSec() int {
	return d.keepTimeoutSec
}

func (d *BuiltInHarborDocker) TraceHeader() bool {
	return d.traceHeader
}

func (d *BuiltInHarborDocker) TourBufferSize() int {
	return d.tourBufferSize
}

func (d *BuiltInHarborDocker) RedirectFile() string {
	return d.redirectFile
}

func (d *BuiltInHarborDocker) ControlPort() int {
	return d.controlPort
}

func (d *BuiltInHarborDocker) GzipComp() bool {
	return d.gzipComp
}

func (d *BuiltInHarborDocker) NetMultiplexer() int {
	return d.netMultiplexer
}

func (d *BuiltInHarborDocker) FileMultiplexer() int {
	return d.fileMultiplexer
}

func (d *BuiltInHarborDocker) LogMultiplexer() int {
	return d.logMultiplexer
}

func (d *BuiltInHarborDocker) CgiMultiplexer() int {
	return d.cgiMultiplexer
}

func (d *BuiltInHarborDocker) Recipient() int {
	return d.recipient
}

func (d *BuiltInHarborDocker) PidFile() string {
	return d.pidFile
}

func (d *BuiltInHarborDocker) MultiCore() bool {
	return d.multiCore
}
