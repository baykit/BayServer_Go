package builtin

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/docker/file"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"
)

type BuiltInCityDocker struct {
	*base.DockerBase

	townList    []docker.Town
	defaultTown docker.Town

	clubList    []docker.Club
	defaultClub docker.Club

	logList        []docker.Log
	permissionList []docker.Permission

	trouble docker.Trouble
	name    string
}

type clubMatchInfo struct {
	club       docker.Club
	scriptName string
	pathInfo   string
}

type matchInfo struct {
	town         docker.Town
	clubMatch    *clubMatchInfo
	queryString  string
	redirectURI  string
	rewrittenURI string
}

func NewBuiltInCityDocker() docker.City {
	c := BuiltInCityDocker{}
	c.DockerBase = base.NewDockerBase(&c)
	c.townList = []docker.Town{}
	c.defaultTown = nil
	c.clubList = []docker.Club{}
	c.defaultClub = nil
	c.logList = []docker.Log{}
	c.permissionList = []docker.Permission{}
	c.trouble = nil
	c.name = ""
	return &c
}

func (c *BuiltInCityDocker) String() string {
	return fmt.Sprintf("City[%s]", c.name)
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (c *BuiltInCityDocker) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	err := c.DockerBase.Init(elm, parent)
	if err != nil {
		return err
	}

	c.name = elm.Arg
	sort.Slice(c.townList, func(i, j int) bool {
		return len(c.townList[i].Name()) > len(c.townList[j].Name())
	})

	for _, t := range c.townList {
		baylog.Debug(baymessage.Get(symbol.MSG_SETTING_UP_TOWN, t.Name(), t.Location()))
	}

	c.defaultTown = NewBuiltInTownDocker()
	c.defaultClub = file.NewFileDocker()

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (c *BuiltInCityDocker) InitDocker(dkr docker.Docker) (bool, exception2.ConfigException) {
	switch dkr.GetType() {
	case "town":
		c.townList = append(c.townList, dkr.(docker.Town))
	case "club":
		c.clubList = append(c.clubList, dkr.(docker.Club))
	case "log":
		c.logList = append(c.logList, dkr.(docker.Log))
	case "permission":
		c.permissionList = append(c.permissionList, dkr.(docker.Permission))
	case "trouble":
		c.trouble = dkr.(docker.Trouble)
	default:
		baylog.Error("Invalid docker in city: type=%s", dkr.GetType())
		return false, nil

	}
	return true, nil
}

func (c *BuiltInCityDocker) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception2.ConfigException) {
	return c.DockerBase.DefaultInitKeyVal(kv)
}

/****************************************/
// Implements City
/****************************************/

/**
 * City name (host name)
 */
func (c *BuiltInCityDocker) Name() string {
	return c.name
}

/**
 * All clubs (not included in town) in this city
 */
func (c *BuiltInCityDocker) Clubs() []docker.Club {
	return c.clubList
}

/**
 * All towns in this city
 */
func (c *BuiltInCityDocker) Towns() []docker.Town {
	return c.townList
}

/**
 * Enter city
 */
func (c *BuiltInCityDocker) Enter(tur tour.Tour) exception2.HttpException {
	baylog.Debug("%s City[%s] Request URI: %s", tur, c.name, tur.Req().Uri())
	tur.SetCity(c)

	var herr exception2.HttpException = nil
catch:
	for { // try catch
		for _, p := range c.permissionList {
			herr = p.TourAdmitted(tur)
			if herr != nil {
				break catch
			}
		}

		mInfo := c.getTownAndClub(tur.Req().Uri())
		if mInfo == nil {
			herr = exception2.NewHttpException(httpstatus.NOT_FOUND, tur.Req().Uri())
			break
		}

		herr = mInfo.town.CheckAdmitted(tur)
		if herr != nil {
			break
		}

		if mInfo.redirectURI != "" {
			herr = exception2.NewMovedTemporarily(mInfo.redirectURI)

		} else {
			baylog.Debug("%s Town[%s] Club[%s]", tur, mInfo.town.Name(), mInfo.clubMatch.club)
			tur.Req().SetQueryString(mInfo.queryString)
			tur.Req().SetScriptName(mInfo.clubMatch.scriptName)

			if mInfo.clubMatch.club.Charset() != "" {
				tur.Req().SetCharset(mInfo.clubMatch.club.Charset())
				tur.Res().SetCharset(mInfo.clubMatch.club.Charset())

			} else {
				tur.Req().SetCharset(bayserver.Harbor().Charset())
				tur.Res().SetCharset(bayserver.Harbor().Charset())
			}

			tur.Req().SetPathInfo(mInfo.clubMatch.pathInfo)
			if tur.Req().PathInfo() != "" && mInfo.clubMatch.club.DecodePathInfo() {
				decodedStr, err := url.QueryUnescape(tur.Req().PathInfo())
				if err != nil {
					baylog.ErrorE(exception.NewExceptionFromError(err), "")
				}
				tur.Req().SetPathInfo(decodedStr)
			}

			if mInfo.rewrittenURI != "" {
				tur.Req().SetRewrittenUri(mInfo.rewrittenURI)
			}

			club := mInfo.clubMatch.club
			tur.SetTown(mInfo.town)
			tur.SetClub(club)
			herr = club.Arrive(tur)
			if herr != nil {
				break
			}
		}

		break
	}

	return herr
}

/**
 * Get trouble base
 */
func (c *BuiltInCityDocker) GetTrouble() docker.Trouble {
	return c.trouble
}

/**
 * Logging
 */
func (c *BuiltInCityDocker) Log(tur tour.Tour) {

}

/****************************************/
/* Private methods                      */
/****************************************/

func (c *BuiltInCityDocker) clubMatches(clubs []docker.Club, relUri string, townName string) *clubMatchInfo {
	mi := clubMatchInfo{}
	var anyd docker.Club = nil

	for _, d := range clubs {
		if d.FileName() == "*" && d.Extension() == "" {
			// Ignore any match club
			anyd = d
			break
		}
	}

	// search for club
	tokens := strings.Split(relUri, "/")
	relScriptName := ""

loop:
	for _, fname := range tokens {
		if relScriptName != "" {
			relScriptName += "/"
		}
		relScriptName += fname

		for _, d := range clubs {
			if d == anyd {
				// Ignore any match club
				continue
			}

			if d.Matches(fname) {
				mi.club = d
				break loop
			}
		}
	}

	if mi.club == nil && anyd != nil {
		mi.club = anyd
	}

	if mi.club == nil {
		return nil
	}

	if townName == "/" && relScriptName == "" {
		mi.scriptName = "/"
		mi.pathInfo = ""

	} else {
		mi.scriptName = townName + relScriptName
		if len(relScriptName) == len(relUri) {
			mi.pathInfo = ""
		} else {
			mi.pathInfo = relUri[len(relScriptName):]
		}
	}

	return &mi
}

func (c *BuiltInCityDocker) getTownAndClub(reqUri string) *matchInfo {
	mi := &matchInfo{}

	uri := reqUri
	pos := strings.Index(uri, "?")
	if pos != -1 {
		mi.queryString = uri[pos+1:]
		uri = uri[:pos]
	}

	for _, t := range c.townList {
		m := t.Matches(uri)
		if m == docker.MATCH_TYPE_NOT_MATCHED {
			continue
		}

		// town matched
		mi.town = t
		if m == docker.MATCH_TYPE_CLOSE {
			mi.redirectURI = uri + "/"
			if mi.queryString != "" {
				mi.redirectURI += mi.queryString
			}
			return mi
		}

		orgUri := uri
		uri = t.Reroute(uri)
		if uri != orgUri {
			mi.rewrittenURI = uri
		}

		rel := uri[len(t.Name()):]

		mi.clubMatch = c.clubMatches(t.Clubs(), rel, t.Name())

		if mi.clubMatch == nil {
			mi.clubMatch = c.clubMatches(c.clubList, rel, t.Name())
		}

		if mi.clubMatch == nil {
			// check index file
			if strings.HasSuffix(uri, "/") && t.WelcomeFile() != "" {
				indexUri := uri + t.WelcomeFile()
				relUri := rel + t.WelcomeFile()
				indexLocation := path.Join(t.Location(), relUri)
				if sysutil.IsFile(indexLocation) {
					if mi.queryString != "" {
						indexUri += "?" + mi.queryString
					}
					m2 := c.getTownAndClub(indexUri)
					if m2 != nil {
						// matched
						m2.redirectURI = indexUri
						return m2
					}
				}
			}

			// default club matches
			mi.clubMatch = &clubMatchInfo{}
			mi.clubMatch.club = c.defaultClub
			mi.clubMatch.scriptName = ""
			mi.clubMatch.pathInfo = ""
		}

		return mi
	}

	return nil
}
