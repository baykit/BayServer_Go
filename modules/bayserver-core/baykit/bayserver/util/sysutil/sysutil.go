package sysutil

import (
	"bayserver-core/baykit/bayserver/util/exception"
	"net"
	"os"
	"time"
)

func IsFile(path string) bool {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func IsDirectory(path string) bool {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

func Exists(path string) bool {
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

func GetFileSize(file *os.File) (int64, exception.IOException) {
	fileInfo, err := file.Stat()
	if err != nil {
		return -1, exception.NewIOException("Cannot get stat %s: %s", file.Name(), err)
	}

	return fileInfo.Size(), nil
}

func GetLocalHostAndIp() (string, string, exception.IOException) {
	host, err := os.Hostname()
	if err != nil {
		return "", "", exception.NewIOExceptionFromError(err)
	}

	ips, err := net.LookupIP(host)
	var ipadr string
	if err != nil {
		return "", "", exception.NewIOExceptionFromError(err)

	} else {
		for _, ip := range ips {
			ipv4 := ip.To4()
			if ipv4 != nil {
				ipadr = net.IP.String(ipv4)
				break
			}
		}
	}

	return host, ipadr, nil
}

func Pid() int {
	return os.Getpid()
}

func CurrentTimeSecs() int64 {
	now := time.Now()
	return now.Unix()
}
