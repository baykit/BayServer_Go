package builtin

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/symbol"
	"strconv"
	"strings"
)

type BuiltInTroubleDocker struct {
	*base.DockerBase

	cmdMap map[int]*docker.TroubleCommand
}

func NewBuiltInTroubleDocker() docker.Trouble {
	t := &BuiltInTroubleDocker{}
	t.DockerBase = base.NewDockerBase(t)
	t.cmdMap = make(map[int]*docker.TroubleCommand)

	// interface check
	var _ docker.Docker = t
	var _ docker.Trouble = t
	return t
}

func (t *BuiltInTroubleDocker) String() string {
	return "BuiltInTroubleDocker"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (t *BuiltInTroubleDocker) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	return t.DockerBase.Init(elm, parent)
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (t *BuiltInTroubleDocker) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	return t.DockerBase.DefaultInitDocker()
}

func (t *BuiltInTroubleDocker) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {
	status, err := strconv.Atoi(kv.Key)
	if err != nil {
		return false, exception2.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PARAMETER, kv.Key))
	}

	pos := strings.Index(kv.Value, " ")
	if pos <= 0 {
		return false, exception2.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
	}

	mstr := strings.ToLower(kv.Value[0:pos])
	var method = 0
	if mstr == "guide" {
		method = docker.TROUBLE_METHOD_GUIDE
	} else if mstr == "text" {
		method = docker.TROUBLE_METHOD_TEXT
	} else if mstr == "reroute" {
		method = docker.TROUBLE_METHOD_REROUTE
	} else {
		return false, exception2.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE, kv.Value))
	}

	t.cmdMap[status] = docker.NewTroubleCommand(method, kv.Value[pos+1:])
	return true, nil
}

/****************************************/
/* Implements Trouble                   */
/****************************************/

func (t *BuiltInTroubleDocker) Find(status int) *docker.TroubleCommand {
	return t.cmdMap[status]
}
