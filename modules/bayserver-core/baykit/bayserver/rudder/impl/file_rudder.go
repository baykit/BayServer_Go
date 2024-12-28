package impl

import (
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util/exception"
	"os"
	"runtime"
)

type FileRudder struct {
	File *os.File
	fd   int
}

func NewFileRudder(file *os.File) *FileRudder {
	rd := FileRudder{
		File: file,
		fd:   0,
	}
	if runtime.GOOS != "windows" {
		rd.fd = int(file.Fd())
	}
	return &rd
}

func (rd *FileRudder) String() string {
	return "File[" + rd.File.Name() + "]"
}

func GetFile(rd rudder.Rudder) *os.File {
	return rd.(*FileRudder).File
}

/****************************************/
/* Implements Rudder                    */
/****************************************/

func (rd *FileRudder) Key() interface{} {
	return rd.File
}

func (rd *FileRudder) Fd() int {
	return rd.fd
}

func (rd *FileRudder) SetNonBlocking() exception.IOException {
	return nil
}

func (rd *FileRudder) Read(buf []byte) (int, exception.IOException) {
	n, err := rd.File.Read(buf)
	if err != nil {
		return n, exception.NewIOExceptionFromError(err)
	} else {
		return n, nil
	}
}

func (rd *FileRudder) Write(buf []byte) (int, exception.IOException) {
	n, err := rd.File.Write(buf)
	if err != nil {
		return n, exception.NewIOExceptionFromError(err)
	} else {
		return n, nil
	}
}

func (rd *FileRudder) Close() exception.IOException {
	err := rd.File.Close()
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	} else {
		return nil
	}
}
