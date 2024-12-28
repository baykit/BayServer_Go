package md5password

import (
	"crypto/md5"
	"encoding/hex"
)

func Encode(password string) string {
	dig := md5.Sum([]byte(password))
	return hex.EncodeToString(dig[:])
}
