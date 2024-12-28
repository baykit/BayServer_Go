//go:build unix

package ioutil

import (
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/util/exception"
	"golang.org/x/sys/unix"
	"os"
)

func OpenLocalPipe() ([]rudder.Rudder, exception.IOException) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		return nil, exception.NewIOExceptionFromError(err)
	} else {
		s1 := os.NewFile(uintptr(fds[0]), "socket1")
		s2 := os.NewFile(uintptr(fds[1]), "socket2")
		return []rudder.Rudder{impl.NewFileRudder(s1), impl.NewFileRudder(s2)}, nil
	}
}
