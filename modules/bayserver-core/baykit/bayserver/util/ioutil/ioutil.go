package ioutil

import (
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/rudder/impl"
	"bayserver-core/baykit/bayserver/util/exception"
	"encoding/binary"
)

func ReadInt32(rd rudder.Rudder) (int32, bool, exception.IOException) {
	buf := make([]byte, 4)

	n, err := rd.Read(buf)
	if err != nil {
		return -1, false, exception.NewIOExceptionFromError(err)
	}
	if n == 0 {
		// EOF
		return -1, true, nil
	}
	if n != 4 {
		return -1, false, exception.NewIOException("No enough data")
	}

	return int32(binary.BigEndian.Uint32(buf)), false, nil
}

func WriteInt32(rd rudder.Rudder, val int32) exception.IOException {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(val))

	_, err := rd.Write(buf)
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	} else {
		return nil
	}
}

func GetSockRecvBufSize(rd rudder.Rudder) (int, exception.IOException) {
	return rd.(*impl.TcpConnRudder).GetSocketReceiveBufferSize()
}
