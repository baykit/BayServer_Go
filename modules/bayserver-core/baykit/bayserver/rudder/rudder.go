package rudder

import (
	"bayserver-core/baykit/bayserver/util/exception"
)

type Rudder interface {
	Key() interface{}
	Read(buf []byte) (int, exception.IOException)
	Write(buf []byte) (int, exception.IOException)
	Close() exception.IOException
}
