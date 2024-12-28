package httputil

import (
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bytes"
	"net"
)

const MAX_LINE_LEN = 5000

/**
 * Read a line from stream
 *
 * @return line as string
 */
func ReadLine(in rudder.Rudder) (string, exception.IOException) {

	/** Current reading line */
	buf := make([]byte, MAX_LINE_LEN)

	c := 0
	var n int
	for n = 0; n < len(buf); n++ {

		var err error
		c, err = in.Read(buf)
		if err != nil {
			return "", exception.NewIOExceptionFromError(err)
		}

		if c == -1 {
			break
		}

		// If character is newline, end to read line
		if c == '\n' {
			break
		}

		// Put the character to buffer
		buf[n] = byte(c)
	}

	if n == 0 && c == -1 {
		return "", nil
	}

	// If line is too long, return error
	if n == len(buf) {
		return "", exception.NewIOException("Request line too long")
	}

	// Remove a '\r' character
	if n != 0 && buf[n-1] == '\r' {
		n--
	}

	// Create line as string
	return string(buf[:n]), nil
}

/**
 * Send MIME headers This method is called from sendHeaders()
 */
func SendMimeHeaders(hdr *headers.Headers, buf *bytes.Buffer) exception.IOException {

	var ioerr exception.IOException = nil
	for _, name := range hdr.HeaderNames() {
		for _, values := range hdr.HeaderValues(name) {
			_, err := buf.Write([]byte(name))
			if err != nil {
				ioerr = exception.NewIOExceptionFromError(err)
				break
			}
			_, err = buf.Write(headers.HEADER_SEPARATOR_BYTES)
			if err != nil {
				ioerr = exception.NewIOExceptionFromError(err)
				break
			}
			_, err = buf.Write([]byte(values))
			if err != nil {
				ioerr = exception.NewIOExceptionFromError(err)
				break
			}
			_, err = buf.Write([]byte("\r\n"))
			if err != nil {
				ioerr = exception.NewIOExceptionFromError(err)
				break
			}
		}
		if ioerr != nil {
			break
		}
	}

	return ioerr
}

func SendNewLine(buf *bytes.Buffer) exception.IOException {
	_, err := buf.Write([]byte("\r\n"))
	if err != nil {
		return exception.NewIOExceptionFromError(err)
	} else {
		return nil
	}
}

func ResolveHost(adr string) (string, exception.IOException) {
	hostNames, err := net.LookupAddr(adr)
	if err != nil {
		return "", exception.NewIOExceptionFromError(err)

	} else {
		if len(hostNames) == 0 {
			return "", nil

		} else {
			return hostNames[0], nil
		}
	}
}
