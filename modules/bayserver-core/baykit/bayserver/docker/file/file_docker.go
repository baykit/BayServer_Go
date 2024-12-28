package file

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/base"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/baylog"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/strutil"
	"bayserver-core/baykit/bayserver/util/sysutil"
	"net/url"
	"path"
	"strings"
)

type FileDocker struct {
	*base.ClubBase
	listFiles bool
}

func NewFileDocker() docker.Club {
	d := &FileDocker{}
	d.ClubBase = base.NewClubBase(d)
	d.listFiles = false

	var _ docker.Club = d // implement check
	return d

}

func (d *FileDocker) String() string {
	return "File"
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *FileDocker) Init(elm *bcf.BcfElement, parent docker.Docker) exception.ConfigException {
	return d.ClubBase.Init(elm, parent)
}

/****************************************/
/* Implements DockerInitializer         */
/****************************************/

func (d *FileDocker) InitKeyVal(kv *bcf.BcfKeyVal) (bool, exception.ConfigException) {

	switch strings.ToLower(kv.Key) {
	case "listfiles":
		var err exception2.Exception
		d.listFiles, err = strutil.ParseBool(kv.Value)
		if err != nil {
			baylog.ErrorE(err, "")
			return false, exception.NewConfigException(
				kv.FileName,
				kv.LineNo,
				baymessage.Get(symbol.CFG_INVALID_PARAMETER_VALUE),
				kv.Value)
		}

	default:
		_, cerr := d.ClubBase.InitKeyVal(kv)
		if cerr != nil {
			return false, cerr
		}
	}

	return true, nil
}

/****************************************/
/* Implements Club                      */
/****************************************/

func (d *FileDocker) Arrive(tur tour.Tour) exception.HttpException {

	var relPath string
	if tur.Req().RewrittenUri() != "" {
		relPath = tur.Req().RewrittenUri()
	} else {
		relPath = tur.Req().Uri()
	}

	town := tur.Town().(docker.Town)
	if town.Name() != "" {
		relPath = relPath[len(town.Name()):]
	}

	pos := strings.Index(relPath, "?")
	if pos != -1 {
		relPath = relPath[:pos]
	}

	relPath, err := url.QueryUnescape(relPath)
	if err != nil {
		baylog.ErrorE(exception2.NewIOExceptionFromError(err), "")
	}

	real := path.Join(tur.Town().(docker.Town).Location(), relPath)

	if sysutil.IsDirectory(real) && d.listFiles {

	} else {
		handler := NewFileContentHandler(real)
		tur.Req().SetReqContentHandler(handler)
	}

	return nil
}
