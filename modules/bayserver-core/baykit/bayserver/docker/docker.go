package docker

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/exception"
)

type Docker interface {
	Init(ini *bcf.BcfElement, parent Docker) exception.ConfigException

	GetType() string
}
