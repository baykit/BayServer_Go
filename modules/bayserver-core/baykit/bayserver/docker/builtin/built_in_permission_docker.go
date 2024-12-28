package builtin

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/common/groups"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"net"
	"strings"
)

/****************************************/
/* type PermissionMatcher               */
/****************************************/

type PermissionMatcher interface {
	matchSocket(rd rudder.Rudder) bool
	matchTour(tour tour.Tour) bool
}

/****************************************/
/* type HostPermissionMatcher           */
/****************************************/

type HostPermissionMatcher struct {
	matcher *util.HostMatcher
}

func NewHostPermissionMatcher(hostPtn string) *HostPermissionMatcher {
	return &HostPermissionMatcher{
		matcher: util.NewHostMatcher(hostPtn),
	}
}

func (m *HostPermissionMatcher) matchSocket(rd rudder.Rudder) bool {
	if tcpRd, ok := rd.(*impl.TcpConnRudder); ok {
		remoteAddr := tcpRd.Conn.RemoteAddr().(*net.TCPAddr)
		hostNames, err := net.LookupHost(remoteAddr.IP.String())
		if err != nil {
			baylog.ErrorE(exception.NewIOExceptionFromError(err), "Lookup error")
			return false
		}
		return m.matcher.Match(hostNames[0])
	} else {
		bayserver.FatalError(exception.NewSink("UPD not supported"))
		return false
	}
}

func (m *HostPermissionMatcher) matchTour(tur tour.Tour) bool {
	return m.matcher.Match(tur.Req().RemoteHost())
}

/****************************************/
/* type HostPermissionMatcher           */
/****************************************/

type IpPermissionMatcher struct {
	matcher *util.IpMatcher
}

func NewIpPermissionMatcher(cidr string) (*IpPermissionMatcher, exception.IOException) {
	m := IpPermissionMatcher{}
	var ioerr exception.IOException
	m.matcher, ioerr = util.NewIpMatcher(cidr)
	if ioerr != nil {
		return nil, ioerr
	}
	return &m, ioerr
}

func (m *IpPermissionMatcher) matchSocket(rd rudder.Rudder) bool {
	if tcpRd, ok := rd.(*impl.TcpConnRudder); ok {
		remoteAddr := tcpRd.Conn.RemoteAddr().(*net.TCPAddr)
		return m.matcher.Match(remoteAddr.IP)
	} else {
		bayserver.FatalError(exception.NewSink("UPD not supported"))
		return false
	}
}

func (m *IpPermissionMatcher) matchTour(tur tour.Tour) bool {
	return m.matcher.Match(net.ParseIP(tur.Req().RemoteAddress()))
}

/****************************************/
/* type CheckItem                       */
/****************************************/

type CheckItem struct {
	matcher PermissionMatcher
	admit   bool
}

func NewCheckItem(matcher PermissionMatcher, admit bool) *CheckItem {
	return &CheckItem{
		matcher: matcher,
		admit:   admit,
	}
}

func (c *CheckItem) socketAdmitted(rd rudder.Rudder) bool {
	return c.matcher.matchSocket(rd) == c.admit
}

func (c *CheckItem) tourAdmitted(tur tour.Tour) bool {
	return c.matcher.matchTour(tur) == c.admit
}

/****************************************/
/* type BuildInPermissionDocker         */
/****************************************/

type BuiltInPermissionDocker struct {
	*base.DockerBase

	checkList []*CheckItem
	groups    []*groups.Group
}

func NewPermissionDocker() docker.Permission {
	d := &BuiltInPermissionDocker{}
	d.DockerBase = base.NewDockerBase(d)
	d.checkList = make([]*CheckItem, 0)
	d.groups = make([]*groups.Group, 0)
	return d
}

func (d *BuiltInPermissionDocker) String() string {
	return "BuiltInPermissionDocker"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *BuiltInPermissionDocker) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	err := d.DockerBase.Init(elm, parent)
	if err != nil {
		return err
	}

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (d *BuiltInPermissionDocker) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	return false, nil
}

func (d *BuiltInPermissionDocker) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {
	var ioerr exception.IOException = nil
	var cerr exception2.ConfigException = nil

catch:
	for {
		// try catch
		switch strings.ToLower(kv.Key) {
		default:
			return d.DockerBase.DefaultInitKeyVal(kv)

		case "admit", "allow":
			var matchers []PermissionMatcher = nil
			matchers, cerr, ioerr = d.parseValue(kv)
			if cerr != nil || ioerr != nil {
				break catch
			}

			for _, pm := range matchers {
				d.checkList = append(d.checkList, NewCheckItem(pm, true))
			}
			break

		case "refuse", "deny":
			var matchers []PermissionMatcher = nil
			matchers, cerr, ioerr = d.parseValue(kv)
			if cerr != nil || ioerr != nil {
				break catch
			}

			for _, pm := range matchers {
				d.checkList = append(d.checkList, NewCheckItem(pm, false))
			}
			break

		case "group":
			tokens := strings.Fields(kv.Value)
			for _, tk := range tokens {
				g := groups.GetGroup(tk)
				if g == nil {
					cerr = exception2.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_GROUP_NOT_FOUND, kv.Value))
					break catch
				}
				d.groups = append(d.groups, g)
			}
			break
		}
		return true, nil
	}

	if cerr != nil {
		return false, cerr

	} else {
		return false, exception2.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PERMISSION_DESCRIPTION, kv.Value))
	}

}

/****************************************/
/* Implements Permission                */
/****************************************/

func (d *BuiltInPermissionDocker) SocketAdmitted(rd rudder.Rudder) exception2.HttpException {
	// Check remote host
	isOk := true
	for _, chk := range d.checkList {
		if chk.admit {
			if chk.socketAdmitted(rd) {
				isOk = true
				break
			}

		} else {
			if !chk.socketAdmitted(rd) {
				isOk = false
				break
			}
		}
	}

	if !isOk {
		baylog.Error("Permission error: socket not admitted: %s", rd)
		return exception2.NewHttpException(httpstatus.FORBIDDEN, "")
	}

	return nil
}

func (d *BuiltInPermissionDocker) TourAdmitted(tur tour.Tour) exception2.HttpException {
	// Check URI
	isOk := true
	for _, chk := range d.checkList {
		if chk.admit {
			if chk.tourAdmitted(tur) {
				isOk = true
				break
			}
		} else {
			if !chk.tourAdmitted(tur) {
				isOk = false
				break
			}
		}
	}

	if !isOk {
		return exception2.NewHttpException(httpstatus.FORBIDDEN, tur.Req().Uri())
	}

	if len(d.groups) == 0 {
		return nil
	}

	// Check member
	isOk = false
	if tur.Req().RemoteUser() != "" {
		for _, g := range d.groups {
			if g.Validate(tur.Req().RemoteUser(), tur.Req().RemotePass()) {
				isOk = true
				break
			}
		}
	}

	if !isOk {
		tur.Res().Headers().Set(headers.WWW_AUTHENTICATE, "Basic realm=\"Auth\"")
		return exception2.NewHttpException(httpstatus.UNAUTHORIZED, "")
	}

	return nil
}

/****************************************/
/* Private methods                      */
/****************************************/

func (d *BuiltInPermissionDocker) parseValue(kv *bcf.BcfKeyVal) ([]PermissionMatcher, exception2.ConfigException, exception.IOException) {
	tokens := strings.Fields(kv.Value)
	typ := ""
	matchStr := make([]string, 0)

	for i, tk := range tokens {
		if i == 0 {
			typ = tk
		} else {
			matchStr = append(matchStr, tk)
		}
	}

	if len(matchStr) == 0 {
		return nil, exception2.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PERMISSION_DESCRIPTION, kv.Value)), nil
	}

	pmList := make([]PermissionMatcher, 0)
	typ = strings.ToLower(typ)
	if typ == "host" {
		for _, m := range matchStr {
			pmList = append(pmList, NewHostPermissionMatcher(m))
		}
		return pmList, nil, nil

	} else if typ == "ip" {
		for _, m := range matchStr {
			mch, ioerr := NewIpPermissionMatcher(m)
			if ioerr != nil {
				return nil, nil, ioerr
			}
			pmList = append(pmList, mch)
		}
		return pmList, nil, nil

	} else {
		return nil, exception2.NewConfigException(kv.FileName, kv.LineNo, baymessage.Get(symbol.CFG_INVALID_PERMISSION_DESCRIPTION, kv.Value)), nil

	}
}
