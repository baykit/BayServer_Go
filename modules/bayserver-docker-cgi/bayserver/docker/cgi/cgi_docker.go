package cgi

import (
	"bayserver-core/baykit/bayserver/docker"
	"bayserver-core/baykit/bayserver/util/cgiutil"
	"bayserver-core/baykit/bayserver/util/exception"
	"os/exec"
)

type CgiDocker struct {
	*CgiDockerBase
	interpreter string
	scriptBase  string
	docRoot     string
	timeoutSec  int
}

func NewCgiDocker() docker.Docker {
	dkr := &CgiDocker{
		interpreter: "",
		scriptBase:  "",
		docRoot:     "",
		timeoutSec:  DEFAULT_TIMEOUT_SEC,
	}
	dkr.CgiDockerBase = NewCgiDockerBase(dkr)

	var _ docker.Club = dkr // implement check

	return dkr
}

func (d *CgiDocker) String() string {
	return "CgiDocker"
}

/****************************************/
/* Implements CgiDockerSub           */
/****************************************/

func (d *CgiDocker) CreateProcess(env map[string]string) (*exec.Cmd, exception.IOException) {
	script := env[cgiutil.SCRIPT_FILENAME]
	var cmd *exec.Cmd
	if d.interpreter == "" {
		cmd = exec.Command(script)
	} else {
		cmd = exec.Command(d.interpreter, script)
	}
	cmd.Env = []string{}
	for name, value := range env {
		cmd.Env = append(cmd.Env, name+"="+value)
	}

	return cmd, nil
}
