package base

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/strutil"
	"strings"
)

type ClubBase struct {
	*DockerBase
	fileName       string
	extension      string
	charset        string
	decodePathInfo bool
}

func NewClubBase(parent DockerInitializer) *ClubBase {
	h := &ClubBase{
		decodePathInfo: true,
	}
	h.DockerBase = NewDockerBase(parent)
	return h
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (c *ClubBase) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	cerr := c.DockerBase.Init(elm, parent)
	if cerr != nil {
		return cerr
	}
	p := strings.LastIndex(elm.Arg, ".")
	if p == -1 {
		c.fileName = elm.Arg
		c.extension = ""
	} else {
		c.fileName = elm.Arg[:p]
		c.extension = elm.Arg[p+1:]
	}

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (c *ClubBase) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	return false, nil
}

func (c *ClubBase) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {
	switch strings.ToLower(kv.Key) {
	default:
		return false, nil

	case "decodepathinfo":
		var err exception.Exception
		c.decodePathInfo, err = strutil.ParseBool(kv.Value)
		if err != nil {
			baylog.ErrorE(err, "")
			return false, exception2.NewConfigException(
				kv.FileName,
				kv.LineNo,
				baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE),
				kv.Value)
		}

	case "charset":
		if kv.Value != "" {
			c.charset = kv.Value
		}
	}
	return true, nil
}

/****************************************/
/* Custom club                     */
/****************************************/

func (c *ClubBase) FileName() string {
	return c.fileName
}

func (c *ClubBase) Extension() string {
	return c.extension
}

func (c *ClubBase) Charset() string {
	return c.charset
}

func (c *ClubBase) DecodePathInfo() bool {
	return c.decodePathInfo
}

/****************************************/
/* Custom functions                     */
/****************************************/

func (c *ClubBase) Matches(fname string) bool {
	// check club
	pos := strings.Index(fname, ".")
	if pos == -1 {
		// fname has no extension
		if c.extension != "" {
			return false
		}

		if c.fileName == "*" {
			return true
		}

		return fname == c.fileName

	} else {
		// fname has extension
		if c.extension == "" {
			return false
		}

		nm := fname[0:pos]
		ext := fname[pos+1:]

		if c.extension != "*" && ext != c.extension {
			return false
		}

		if c.fileName == "*" {
			return true
		} else {
			return nm == c.fileName
		}
	}
}
