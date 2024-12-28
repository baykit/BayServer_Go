//go:build windows

package impl

import (
	"syscall"
)

func GetSockOpt(fd, level, opt int) (int, error) {
	return syscall.GetsockoptInt(syscall.Handle(fd), level, opt)
}
