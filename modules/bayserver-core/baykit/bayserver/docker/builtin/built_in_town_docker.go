package builtin

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"path/filepath"
	"strings"
)

type BuiltInTownDocker struct {
	*base.DockerBase

	location       string
	welcome        string
	clubList       []docker.Club
	permissionList []docker.Permission
	rerouteList    []docker.Reroute
	city           docker.City
	name           string
}

func NewBuiltInTownDocker() docker.Town {
	t := BuiltInTownDocker{}
	t.DockerBase = base.NewDockerBase(&t)
	t.location = ""
	t.welcome = ""
	t.clubList = []docker.Club{}
	t.permissionList = []docker.Permission{}
	t.rerouteList = []docker.Reroute{}
	t.city = nil
	t.name = ""
	return &t
}

func (t *BuiltInTownDocker) String() string {
	return "BuiltInTownDocker(name=" + t.name + ",loc=" + t.location + ")"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (t *BuiltInTownDocker) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	arg := elm.Arg
	if !strings.HasPrefix(arg, "/") {
		arg = "/" + arg
	}
	t.name = arg

	if !strings.HasSuffix(t.name, "/") {
		t.name += "/"
	}
	t.city = parent.(docker.City)

	return t.DockerBase.Init(elm, parent)
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (t *BuiltInTownDocker) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	switch dkr.(type) {
	case docker.Club:
		t.clubList = append(t.clubList, dkr.(docker.Club))

	case docker.Permission:
		t.permissionList = append(t.permissionList, dkr.(docker.Permission))

	case docker.Reroute:
		t.rerouteList = append(t.rerouteList, dkr.(docker.Reroute))

	default:
		return t.DockerBase.DefaultInitDocker()
	}

	return true, nil
}

func (t *BuiltInTownDocker) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {
	switch strings.ToLower(kv.Key) {
	default:
		return t.DockerBase.DefaultInitKeyVal(kv)

	case "location":
		t.location = kv.Value
		if !filepath.IsAbs(t.location) {
			t.location = bayserver.GetLocation(t.location)
		}
		if !sysutil.IsDirectory(t.location) {
			return false, exception2.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_LOCATION, t.location))
		}
		break

	case "index", "welcome":
		t.welcome = kv.Value
	}
	return true, nil
}

/****************************************/
/* Implements Town                */
/****************************************/

func (t *BuiltInTownDocker) Name() string {
	return t.name
}

func (t *BuiltInTownDocker) City() docker.City {
	return t.city
}

func (t *BuiltInTownDocker) Location() string {
	return t.location
}

func (t *BuiltInTownDocker) WelcomeFile() string {
	return t.welcome
}

func (t *BuiltInTownDocker) Clubs() []docker.Club {
	return t.clubList
}

func (t *BuiltInTownDocker) Reroute(uri string) string {
	for _, r := range t.rerouteList {
		uri = r.Reroute(t, uri)
	}

	return uri
}

func (t *BuiltInTownDocker) Matches(uri string) int {
	if strings.HasPrefix(uri, t.name) {
		return docker.MATCH_TYPE_MATCHED
	} else if uri+"/" == t.name {
		return docker.MATCH_TYPE_CLOSE
	} else {
		return docker.MATCH_TYPE_NOT_MATCHED
	}
}

func (t *BuiltInTownDocker) CheckAdmitted(tur tour.Tour) exception2.HttpException {
	for _, p := range t.permissionList {
		hterr := p.TourAdmitted(tur)
		if hterr != nil {
			return hterr
		}
	}
	return nil
}
