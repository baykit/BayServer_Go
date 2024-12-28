package base

import (
	"bayserver-core/baykit/bayserver/bcf"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
)

type RerouteBase struct {
	*DockerBase
}

func NewRerouteBase(d DockerInitializer) *RerouteBase {
	p := &RerouteBase{}
	p.DockerBase = NewDockerBase(d)
	return p
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (p *RerouteBase) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {

	name := elm.Arg
	if name != "*" {
		return exception2.NewConfigException(
			elm.FileName,
			elm.LineNo,
			"Invalid reroute name: "+name)
	}

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (p *RerouteBase) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	return p.DefaultInitDocker()
}

func (p *RerouteBase) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {
	return p.DockerBase.DefaultInitKeyVal(kv)
}

/****************************************/
/* Custom methods                       */
/****************************************/

func (p *RerouteBase) Match(uri string) bool {
	return true
}
