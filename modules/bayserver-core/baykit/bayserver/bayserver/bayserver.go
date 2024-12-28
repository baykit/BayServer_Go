package bayserver

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/common"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
)

const ENV_BAYSERVER_HOME = "BSERV_HOME"
const ENV_BAYSERVER_PLAN = "BSERV_PLAN"
const ENV_BAYSERVER_LIB = "BSERV_LIB"

// BayServer home directory
var BservHome func() string

// BayServer plan file (full path)
var BservPlan func() string

// BayServer plan directory (full path)
var PlanDir func() string

// Library directory
var LibDir func() string

var Harbor func() docker.Harbor

var AnchorablePortMap func() map[rudder.Rudder]docker.Port

var UnnchorablePortMap func() map[rudder.Rudder]docker.Port

var Ports func() []docker.Port

var Cities func() *common.Cities

// Main
var Main func(args []string)

var GetLocation func(location string) string

// Load BCF file
var LoadBcf func(fileName string) (*bcf.BcfDocument, bcf.ParseException)

// Load message file
var LoadMessage func(
	filePrefix string,
	locale *util.Locale) (util.Message, bcf.ParseException)

var FindCity func(name string) docker.City

var ParsePath func(location string) (string, exception.BayException)

var SoftwareName func() string

var FatalError func(err exception2.Exception)

var BDefer func()
