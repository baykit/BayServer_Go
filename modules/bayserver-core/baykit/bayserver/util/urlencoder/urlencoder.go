package urlencoder

import "bytes"

/**
 * Encode tilde char only
 */

func EncodeTilde(url string) string {
	buf := bytes.NewBuffer(make([]byte, 0, len(url)))
	for _, c := range url {
		if c == '~' {
			buf.WriteString("%7E")
		} else {
			buf.WriteByte(byte(c))
		}
	}
	return buf.String()
}
