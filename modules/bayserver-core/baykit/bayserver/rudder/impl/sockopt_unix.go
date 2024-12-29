//go:build unix

package impl

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"syscall"
)

func getSockOptInt(fd int, level, opt int) (int, exception.IOException) {
	val, err := syscall.GetsockoptInt(fd, level, opt)
	if err != nil {
		return 0, exception.NewIOExceptionFromError(err)
	}

	return val, nil
}
