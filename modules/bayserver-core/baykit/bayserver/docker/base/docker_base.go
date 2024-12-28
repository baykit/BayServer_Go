package base

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baydockers"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"fmt"
	"strings"
)

type DockerInitializer interface {
	fmt.Stringer
	InitDocker(dkr docker.Docker) (bool, exception.ConfigException)
	InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception.ConfigException)
}

type DockerBase struct {
	parent DockerInitializer
	typ    string
}

func NewDockerBase(parent DockerInitializer) *DockerBase {
	d := DockerBase{}
	d.typ = ""
	d.parent = parent
	return &d
}

func (d *DockerBase) String() string {
	return d.parent.String()
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *DockerBase) Init(elm *bcf.BcfElement, parent docker.Docker) exception.ConfigException {
	d.typ = elm.Name
	baylog.Debug("%s SetType: %s", d, d.typ)
	for _, o := range elm.ContentList {
		if kv, ok := o.(*bcf.BcfKeyVal); ok {
			ok, cerr := d.parent.InitKeyVal(kv)
			if cerr != nil {
				return cerr

			} else if !ok {
				return exception.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PARAMETER, kv.Key))

			}

		} else {
			element, _ := o.(*bcf.BcfElement)

			dkr, ex := baydockers.CreateDockerByElement(element, d.parent.(docker.Docker))
			if ex != nil {
				baylog.ErrorE(ex, "")
				return exception.NewConfigException(
					element.FileName,
					element.LineNo,
					ex.Error())
			}

			ok, cerr := d.parent.InitDocker(dkr)

			if cerr != nil {
				return cerr
			}
			if !ok {
				return exception.NewConfigException(
					element.FileName,
					element.LineNo,
					baymessage.Get(symbol.CFG_INVALID_DOCKER, element.Name))
			}
		}
	}
	return nil
}

/****************************************/
/* Base methods                         */
/****************************************/

func (d *DockerBase) GetType() string {
	baylog.Debug("%s GetType: %s", d, d.typ)
	return d.typ
}

func (d *DockerBase) DefaultInitDocker() (bool, exception.ConfigException) {
	return false, nil
}

func (d *DockerBase) DefaultInitKeyVal(kv *bcf.BcfKeyVal) (bool, exception.ConfigException) {
	switch strings.ToLower(kv.Key) {
	default:
		return false, nil

	case "docker":
		return true, nil
	}
}
