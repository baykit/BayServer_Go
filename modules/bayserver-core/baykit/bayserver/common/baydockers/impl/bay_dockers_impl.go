package impl

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baydockers"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/docker/builtin"
	"bayserver-core/baykit/bayserver/docker/file"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/baylog"
	ajpimpl "bayserver-docker-ajp/baykit/bayserver/docker/ajp/impl"
	"bayserver-docker-cgi/bayserver/docker/cgi"
	fcgiimpl "bayserver-docker-fcgi/baykit/bayserver/docker/fcgi/impl"
	httpimpl "bayserver-docker-http/baykit/bayserver/docker/http/impl"
	wpimpl "bayserver-docker-wordpress/baykit/bayserver/docker/wordpress/impl"
)

var dockerMap map[string]func() docker.Docker

func Init(m map[string]string) exception.ConfigException {
	dockerMap = map[string]func() docker.Docker{}
	baydockers.CreateDocker = CreateDocker
	baydockers.CreateDockerByElement = CreateDockerByElement

	for key, val := range m {
		var f func() docker.Docker
		switch val {
		case "baykit.bayserver.docker.builtin.BuiltInHarborDocker":
			f = func() docker.Docker { return builtin.NewBuiltInHarborDocker() }

		case "baykit.bayserver.docker.http.HtpPortDocker":
			f = func() docker.Docker { return httpimpl.NewHtpPort() }

		case "baykit.bayserver.docker.ajp.AjpPortDocker":
			f = func() docker.Docker { return ajpimpl.NewAjpPort() }

		case "baykit.bayserver.docker.fcgi.FcgPortDocker":
			f = func() docker.Docker { return fcgiimpl.NewFcgPort() }

		case "baykit.bayserver.docker.h3.H3PortDocker":
			f = func() docker.Docker { return nil }

		case "baykit.bayserver.docker.builtin.BuiltInCityDocker":
			f = func() docker.Docker { return builtin.NewBuiltInCityDocker() }

		case "baykit.bayserver.docker.builtin.BuiltInTownDocker":
			f = func() docker.Docker { return builtin.NewBuiltInTownDocker() }

		case "baykit.bayserver.docker.file.FileDocker":
			f = func() docker.Docker { return file.NewFileDocker() }

		case "baykit.bayserver.docker.cgi.CgiDocker":
			f = func() docker.Docker { return cgi.NewCgiDocker() }

		case "baykit.bayserver.docker.cgi.PhpCgiDocker":
			f = func() docker.Docker { return cgi.NewPhpCgiDocker() }

		case "baykit.bayserver.docker.ajp.AjpWarpDocker":
			f = func() docker.Docker { return ajpimpl.NewAjpWarpDocker() }

		case "baykit.bayserver.docker.fcgi.FcgWarpDocker":
			f = func() docker.Docker { return fcgiimpl.NewFcgWarpDocker() }

		case "baykit.bayserver.docker.http.HtpWarpDocker":
			f = func() docker.Docker { return httpimpl.NewHtpWarpDocker() }

		case "baykit.bayserver.docker.builtin.BuiltInLogDocker":
			f = func() docker.Docker { return builtin.NewBuiltInLogDocker() }

		case "baykit.bayserver.docker.builtin.BuiltInPermissionDocker":
			f = func() docker.Docker { return builtin.NewPermissionDocker() }

		case "baykit.bayserver.docker.builtin.BuiltInSecureDocker":
			f = func() docker.Docker { return builtin.NewBuiltInSecureDocker() }

		case "baykit.bayserver.docker.builtin.BuiltInTroubleDocker":
			f = func() docker.Docker { return builtin.NewBuiltInTroubleDocker() }

		case "baykit.bayserver.docker.wordpress.WordPressDocker":
			f = func() docker.Docker { return wpimpl.NewWordPressDocker() }

		default:
			f = func() docker.Docker { return nil }
		}

		if f == nil {
			baylog.Error("Unkown base: %s", val)
		}
		dockerMap[key] = f
	}

	return nil
}

func CreateDockerByElement(elm *bcf.BcfElement, parent docker.Docker) (docker.Docker, exception.BayException) {
	alias := elm.GetValue("docker")
	baylog.Info("Creating docker: %s %s", elm.Name, alias)
	d, ex := CreateDocker(elm.Name, alias)
	if ex != nil {
		return nil, ex
	}
	if d == nil {
		return nil, exception.NewConfigException(elm.FileName, elm.LineNo, symbol.CFG_INVALID_DOCKER, elm.Name)
	}
	ex = d.Init(elm, parent)
	if ex != nil {
		return nil, ex
	}
	return d, nil
}

func CreateDocker(category string, alias string) (docker.Docker, exception.BayException) {
	var key string
	if alias == "" {
		key = category
	} else {
		key = category + ":" + alias
	}
	factory := dockerMap[key]
	if factory == nil {
		return nil, exception.NewBayException(baymessage.Get(symbol.CFG_DOCKER_NOT_FOUND, key))
	}

	return factory(), nil
	//return (*factory).CreateDocker(), nil
}
