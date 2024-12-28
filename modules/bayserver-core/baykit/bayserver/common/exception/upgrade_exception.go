package exception

import "bayserver-core/baykit/bayserver/util/exception"

type UpgradeException interface {
	exception.IOException

	UpgradeExceptionDummy()
}

type UpgradeExceptionImpl struct {
	exception.IOExceptionImpl
}

// Dummy function
func (e *UpgradeExceptionImpl) UpgradeExceptionDummy() {
}

func NewUpgradeException() UpgradeException {
	ex := UpgradeExceptionImpl{}
	ex.ConstructException(4, nil, "")
	return &ex
}
