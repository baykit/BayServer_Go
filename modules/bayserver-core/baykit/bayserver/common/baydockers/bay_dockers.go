package baydockers

import (
	"bayserver-core/baykit/bayserver/bcf"
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/docker"
)

var CreateDockerByElement func(elm *bcf.BcfElement, parent docker.Docker) (docker.Docker, exception2.BayException)

var CreateDocker func(category string, alias string) (docker.Docker, exception2.BayException)
