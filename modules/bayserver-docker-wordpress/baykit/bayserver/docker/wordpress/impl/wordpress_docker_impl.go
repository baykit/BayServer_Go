package wpimpl

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"path"
	"strings"
)

type WordPressDockerImpl struct {
	*base.RerouteBase
	townPath string
}

func NewWordPressDocker() docker.Reroute {
	h := WordPressDockerImpl{}
	h.RerouteBase = base.NewRerouteBase(&h)
	return &h
}

func (d *WordPressDockerImpl) String() string {
	return "WordPressDocker(townPath=" + d.townPath + ")"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *WordPressDockerImpl) Init(elm *bcf.BcfElement, parent docker.Docker) exception.ConfigException {
	err := d.RerouteBase.Init(elm, parent)
	if err != nil {
		return err
	}

	twn := parent.(docker.Town)
	d.townPath = twn.Location()

	return nil
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

/****************************************/
/* Implements Reroute                      */
/****************************************/

func (d *WordPressDockerImpl) Reroute(twn docker.Town, uri string) string {
	parts := strings.Split(uri, "?")
	uri2 := parts[0]
	if !d.Match(uri2) {
		return uri
	}

	relPath := uri2[len(twn.Name()):]
	if strings.HasPrefix(relPath, "/") {
		relPath = relPath[1:]
	}

	parts = strings.Split(relPath, "/")
	checkPath := ""
	for _, token := range parts {
		if checkPath != "" {
			checkPath += "/"
		}
		checkPath += token
		if sysutil.Exists(path.Join(twn.Location(), checkPath)) {
			return uri
		}
	}

	if !sysutil.Exists(path.Join(twn.Location(), relPath)) {
		return twn.Name() + "index.php/" + uri[len(twn.Name()):]
	} else {
		return uri
	}
}
