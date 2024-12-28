package httpstatus

import (
	"strconv"
)

const OK int = 200
const MOVED_PERMANENTLY int = 301
const MOVED_TEMPORARILY int = 302
const NOT_MODIFIED int = 304
const BAD_REQUEST int = 400
const UNAUTHORIZED int = 401
const FORBIDDEN int = 403
const NOT_FOUND int = 404
const UPGRADE_REQUIRED int = 426
const INTERNAL_SERVER_ERROR int = 500
const SERVICE_UNAVAILABLE int = 503
const GATEWAY_TIMEOUT int = 504
const HTTP_VERSION_NOT_SUPPORTED int = 505

var statusMap = map[int]string{}

func GetDescription(statusCode int) string {
	desc := statusMap[statusCode]
	if desc == "" {
		return "Unknown Status"
	}
	return desc
}

func SetStatusMap(m map[string]string) {
	for key, val := range m {
		code, _ := strconv.Atoi(key)
		statusMap[code] = val
	}
}
