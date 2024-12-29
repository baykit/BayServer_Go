package impl

import (
	bayserver2 "bayserver-core/baykit/bayserver"
	"bayserver-core/baykit/bayserver/agent"
	agentimpl "bayserver-core/baykit/bayserver/agent/impl"
	"bayserver-core/baykit/bayserver/agent/monitor"
	"bayserver-core/baykit/bayserver/agent/signal"
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	bcfimpl "bayserver-core/baykit/bayserver/bcf/impl"
	"bayserver-core/baykit/bayserver/common/baydockers"
	"bayserver-core/baykit/bayserver/common/baydockers/impl"
	"bayserver-core/baykit/bayserver/common/baymessage"
	baymessageimpl "bayserver-core/baykit/bayserver/common/baymessage/impl"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/inboundship/inboundshipstore"
	memusageimpl "bayserver-core/baykit/bayserver/common/memusage/impl"
	"bayserver-core/baykit/bayserver/docker"
	common3 "bayserver-core/baykit/bayserver/docker/common"
	"bayserver-core/baykit/bayserver/protocol/packetstore"
	"bayserver-core/baykit/bayserver/protocol/protocolhandlerstore"
	"bayserver-core/baykit/bayserver/rudder"
	rudderimpl "bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour/tourstore"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"bayserver-core/baykit/bayserver/util/mimes"
	"bayserver-core/baykit/bayserver/util/strutil"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"embed"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var bservHome string /** BayServer home directory */
var bservPlan string /** Configuration file name (full path) */
var planDir string   /** Configuration file directory name (full path) */
var libDir string
var myHostName string   /** Host name */
var myHostAddr string   /** Host address */
var softwareName string /** Software name */

var ports []docker.Port
var harbor docker.Harbor
var anchorablePortMap map[rudder.Rudder]docker.Port
var unanchorablePortMap map[rudder.Rudder]docker.Port
var cities common3.Cities
var logFile *os.File

var resource embed.FS

func Init(res embed.FS) {
	resource = res
	bservHome = ""
	bservPlan = ""
	planDir = ""
	libDir = ""
	myHostName = ""
	myHostAddr = ""
	softwareName = ""
	ports = []docker.Port{}
	harbor = nil
	anchorablePortMap = map[rudder.Rudder]docker.Port{}
	unanchorablePortMap = map[rudder.Rudder]docker.Port{}
	cities = common3.NewCities()

	bayserver.BservHome = func() string { return bservHome }
	bayserver.BservPlan = func() string { return bservPlan }
	bayserver.PlanDir = func() string { return planDir }
	bayserver.LibDir = func() string { return libDir }
	bayserver.Harbor = func() docker.Harbor { return harbor }
	bayserver.AnchorablePortMap = func() map[rudder.Rudder]docker.Port { return anchorablePortMap }
	bayserver.UnnchorablePortMap = func() map[rudder.Rudder]docker.Port { return unanchorablePortMap }
	bayserver.Ports = func() []docker.Port { return ports }
	bayserver.Cities = func() *common3.Cities { return &cities }
	bayserver.Main = Main
	bayserver.GetLocation = GetLocation
	bayserver.LoadMessage = LoadMessage
	bayserver.LoadBcf = LoadBcf
	bayserver.FindCity = FindCity
	bayserver.ParsePath = ParsePath
	bayserver.SoftwareName = SoftwareName
	bayserver.FatalError = FatalError
	bayserver.BDefer = BDefer
}

func Main(args []string) {
	//os_signal.Ignore(syscall.SIGPIPE)

	var cmd = ""
	var home = os.Getenv(bayserver.ENV_BAYSERVER_HOME)
	var plan = os.Getenv(bayserver.ENV_BAYSERVER_PLAN)
	var mkpass = ""
	var init = false

	for _, arg := range args {
		if strings.EqualFold(arg, "-start") {
			cmd = ""

		} else if strings.EqualFold(arg, "-stop") || strings.EqualFold(arg, "-shutdown") {
			cmd = signal.SIGN_AGENT_COMMAND_SHUTDOWN

		} else if strings.EqualFold(arg, "-restartAgents") {
			cmd = signal.SIGN_AGENT_COMMAND_RESTART_AGENTS

		} else if strings.EqualFold(arg, "-reloadCert") {
			cmd = signal.SIGN_AGENT_COMMAND_RELOAD_CERT

		} else if strings.EqualFold(arg, "-memUsage") {
			cmd = signal.SIGN_AGENT_COMMAND_MEM_USAGE

		} else if strings.EqualFold(arg, "-abort") {
			cmd = signal.SIGN_AGENT_COMMAND_ABORT

		} else if strings.EqualFold(arg, "-init") {
			init = true

		} else if strutil.StartsWith(strings.ToLower(arg), "-home=") {
			home = arg[6:]

		} else if strutil.StartsWith(strings.ToLower(arg), "-plan=") {
			plan = arg[6:]

		} else if strutil.StartsWith(strings.ToLower(arg), "-mkpass=") {
			mkpass = arg[8:]

		} else if strutil.StartsWith(strings.ToLower(arg), "-loglevel=") {
			baylog.SetLogLevel(arg[10:])
		}
	}

	if mkpass != "" {
		return
	}

	var err exception.Exception = nil
	for { // try catch
		err = getHome(home)
		if err != nil {
			break
		}
		getLib()

		if init {
			initServer()

		} else {
			err = baymessageimpl.Init()
			if err != nil {
				break
			}

			err = getPlan(plan)
			if err != nil {
				break
			}

			if cmd == "" {
				err = start()

			} else {
				err = signal.SendSign(cmd)
			}
			if err != nil {
				break
			}
		}
		cmd = cmd
		mkpass = mkpass

		break
	}
	if err != nil {
		baylog.ErrorE(err, "")
	}
}

func LoadMessage(
	filePrefix string,
	locale *util.Locale) (util.Message, bcf.ParseException) {

	lang := locale.Language
	fileName := filePrefix + ".bcf"
	if lang != "en" {
		fileName = filePrefix + "_" + lang + ".bcf"
	}

	p := bcfimpl.NewBcfParser()
	doc, ex := p.ParseResource(&resource, fileName)

	if ex != nil {
		return nil, ex
	}

	m := util.NewMessage()
	for _, o := range doc.ContentList {
		if kv, ok := o.(*bcf.BcfKeyVal); ok {
			m.MessageMap()[kv.Key] = kv.Value
		}
	}

	return m, nil
}

func LoadBcf(fileName string) (*bcf.BcfDocument, bcf.ParseException) {
	p := bcfimpl.NewBcfParser()
	doc, ex := p.ParseResource(&resource, fileName)
	if ex != nil {
		return nil, ex
	}
	return doc, nil
}

func FindCity(name string) docker.City {
	return cities.FindCity(name)
}

func ParsePath(location string) (string, exception2.BayException) {
	location = GetLocation(location)

	if !sysutil.IsFile(location) {
		return "", exception2.NewBayException("File not found: %s", location)

	} else {
		return location, nil
	}
}

func GetLocation(location string) string {
	if !filepath.IsAbs(location) {
		return bservHome + string(filepath.Separator) + location

	} else {
		return location
	}
}

func SoftwareName() string {
	if softwareName == "" {
		softwareName = "BayServer/" + getVersion()
	}
	return softwareName
}

func FatalError(err exception.Exception) {
	baylog.FatalE(err, "")
	panic("ERR")
}

func BDefer() {
	if r := recover(); r != nil {
		baylog.FatalE(exception.NewSink("Panic: %v\n", r), "")
		if logFile != nil {
			logFile.Sync()
			logFile.Close()
		}
	}
}

/****************************************/
/* private functions                    */
/****************************************/

func getHome(home string) exception2.BayException {
	// Get BayServer home
	if home == "" {
		home = "."
	}

	var err error
	bservHome, err = filepath.Abs(home)
	if err != nil {
		return exception2.NewBayException("BayServer home is not a directory: %s", bservHome)
	}

	if strutil.EndsWith(bservHome, string(filepath.Separator)) {
		bservHome = bservHome[:len(bservHome)-1]
	}

	baylog.Debug("BayServer home: %s", bservHome)
	return nil
}

func getPlan(plan string) exception2.BayException {
	// Get plan file
	if plan == "" {
		plan = "plan/bayserver.plan"
	}

	if filepath.IsAbs(plan) {
		plan = bservHome + "/" + plan
	}

	bservPlan, _ = filepath.Abs(plan)

	baylog.Debug("BayServer Plan: %s", bservPlan)

	_, err := os.Stat(bservPlan)

	if err == nil {
		// Plan file exists
		fileInfo, _ := os.Stat(bservPlan)
		if fileInfo.IsDir() {
			return exception2.NewBayException("Plan file is not a file: " + bservPlan)
		}

	} else if os.IsNotExist(err) {
		// Plan file does not exist
		return exception2.NewBayException("Plan file not exists: " + bservPlan)

	} else {
		return exception2.NewBayException("Unknown error of plan: " + bservPlan)
	}

	planDir = filepath.Join(bservHome, "plan")
	return nil
}

func getLib() {
	libDir = os.Getenv(bayserver.ENV_BAYSERVER_LIB)
	libDir = filepath.Join(bservHome, "lib")

	baylog.Debug("BayServer Lib: %s", libDir)
}

func initServer() {

}

func start() exception2.BayException {
	var ex exception2.BayException

	for { // try catch
		var dkrDoc *bcf.BcfDocument
		dkrDoc, ex = LoadBcf("resources/conf/dockers.bcf")
		if ex != nil {
			break
		}
		impl.Init(bcfToMap(dkrDoc))

		var mimesDoc *bcf.BcfDocument
		mimesDoc, ex = LoadBcf("resources/conf/mimes.bcf")
		if ex != nil {
			break
		}
		mimes.Init(bcfToMap(mimesDoc))

		var statusDoc *bcf.BcfDocument
		statusDoc, ex = LoadBcf("resources/conf/httpstatus.bcf")
		if ex != nil {
			break
		}
		httpstatus.SetStatusMap(bcfToMap(statusDoc))

		agentimpl.Init()

		ex = loadPlan(bservPlan)
		if ex != nil {
			break
		}

		redirectFile := bayserver.Harbor().RedirectFile()
		if redirectFile != "" {
			if !path.IsAbs(redirectFile) {
				redirectFile = bservHome + "/" + redirectFile
			}

			var err error
			logFile, err = os.Create(redirectFile)
			if err != nil {
				ex = exception.NewIOExceptionFromError(err)
				break
			}
			os.Stdout = logFile
			os.Stderr = logFile
		}
		printVersion()

		if len(ports) == 0 {
			ex = exception2.NewBayException(baymessage.Get(symbol.CFG_NO_PORT_DOCKER))
			break
		}

		var ioerr exception.IOException
		myHostName, myHostAddr, ioerr = sysutil.GetLocalHostAndIp()
		if ioerr != nil {
			baylog.ErrorE(ioerr, "")
		}

		baylog.Debug("Host name    : " + myHostName)
		baylog.Debug("Host address : " + myHostAddr)

		/** Init stores, memory usage managers */
		packetstore.Init()
		inboundshipstore.Init()
		protocolhandlerstore.Init()
		tourstore.Init(tourstore.MAX_TOURS)
		memusageimpl.Init()

		ex = openPorts()
		if ex != nil {
			break
		}

		agent.Init(arrayutil.MakeSequence(1, harbor.GrandAgents()), harbor.MaxShips())

		//invokeRunners//

		ex = monitor.GrandAgentMonitorInit(harbor.GrandAgents())
		if ex != nil {
			break
		}

		signal.SignalAgentInit(harbor.ControlPort())

		ex = createPidFile(sysutil.Pid())
		if ex != nil {
			break
		}

		break
	}

	if ex != nil {
		return ex
	}

	defer func() {
		BDefer()
	}()

	dummy := make(chan bool)
	<-dummy
	return nil
}

func bcfToMap(doc *bcf.BcfDocument) map[string]string {
	m := make(map[string]string)
	for _, o := range doc.ContentList {
		if kv, ok := o.(*bcf.BcfKeyVal); ok {
			m[kv.Key] = kv.Value
		}
	}
	return m
}

func loadPlan(plan string) exception2.BayException {
	p := bcfimpl.NewBcfParser()
	doc, ex := p.Parse(plan)
	if ex != nil {
		return ex
	}
	for _, o := range doc.ContentList {
		if elm, ok := o.(*bcf.BcfElement); ok {
			dkr, ex := baydockers.CreateDockerByElement(elm, nil)
			if ex != nil {
				return ex
			}

			if port, ok := dkr.(docker.Port); ok {
				ports = append(ports, port)

			} else if hb, ok := dkr.(docker.Harbor); ok {
				harbor = hb

			} else if city, ok := dkr.(docker.City); ok {
				cities.Add(city)
			}
		}
	}
	return nil
}

/**
 * Print version information
 */
func printVersion() {

	version := "Version " + getVersion()
	for len(version) < 28 {
		version = " " + version

	}

	println("        ----------------------")
	println("       /     BayServer        \\")
	println("-----------------------------------------------------")
	print(" \\")
	for i := 0; i < 47-len(version); i++ {
		print(" ")
	}
	println(version + "  /")
	println("  \\           Copyright (C) 2000 Yokohama Baykit  /")
	println("   \\                     http://baykit.yokohama  /")
	println("    ---------------------------------------------")
}

/**
 * Get the BayServer version
 */
func getVersion() string {
	return bayserver2.VERSION
}

func openPorts() exception.IOException {

	for _, portDkr := range ports {
		// Open TCP Port
		adr, err := portDkr.Address()
		if err != nil {
			return err
		}
		baylog.Info(adr)

		if portDkr.Anchored() {
			baylog.Info(baymessage.Get(
				symbol.MSG_OPENING_TCP_PORT,
				portDkr.Host(),
				portDkr.PortNo(),
				portDkr.Protocol()))

			server, err := net.Listen("tcp", ":"+strconv.Itoa(portDkr.PortNo()))
			if err != nil {
				return exception.NewIOException(err.Error())
			}

			rd := rudderimpl.NewListenerRudder(server.(*net.TCPListener))
			baylog.Debug("Server rd=%v (fd=%d)", rd, rd.Fd())

			anchorablePortMap[rd] = portDkr
		}
	}

	return nil
}

func createPidFile(pid int) exception.IOException {
	f, err := os.Create(harbor.PidFile())
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	}

	defer f.Close()

	_, err = f.WriteString(strconv.Itoa(pid))
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	}

	return nil
}
