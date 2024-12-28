package cgi

import (
	"bayserver-core/baykit/bayserver/bcf"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/cgiutil"
	"bayserver-core/baykit/bayserver/util/exception"
	"os/exec"
)

type PhpCgiDocker struct {
	*CgiDockerBase
}

func NewPhpCgiDocker() docker.Docker {
	dkr := &PhpCgiDocker{}
	dkr.CgiDockerBase = NewCgiDockerBase(dkr)

	var _ docker.Club = dkr
	var _ docker.Docker = dkr
	return dkr
}

/****************************************/
/* Implements Docker                    */
/****************************************/

func (d *PhpCgiDocker) Init(elm *bcf.BcfElement, parent docker.Docker) exception2.ConfigException {
	err := d.CgiDockerBase.Init(elm, parent)
	if err != nil {
		return err
	}

	if d.interpreter == "" {
		d.interpreter = "php-cgi"
	}

	baylog.Info("PHP interpreter: %s", d.interpreter)

	return nil
}

/****************************************/
/* Implements PhpCgiDockerSub           */
/****************************************/

func (d *PhpCgiDocker) CreateProcess(env map[string]string) (*exec.Cmd, exception.IOException) {
	env["PHP_SELF"] = env[cgiutil.SCRIPT_FILENAME]
	env["REDIRECT_STATUS"] = "200"

	var cmd *exec.Cmd
	cmd = exec.Command(d.interpreter)
	cmd.Env = []string{}
	for name, value := range env {
		cmd.Env = append(cmd.Env, name+"="+value)
	}

	return cmd, nil
}
