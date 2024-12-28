package headers

import (
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/httpstatus"
	"strconv"
	"strings"
)

/**
 * Known header names
 */
const HEADER_SEPARATOR = ": "

var HEADER_SEPARATOR_BYTES = []byte(HEADER_SEPARATOR)

const CONTENT_TYPE = "content-type"
const CONTENT_LENGTH = "content-length"
const CONTENT_ENCODING = "content-encoding"
const HDR_TRANSFER_ENCODING = "Transfer-Encoding"
const CONNECTION = "Connection"
const AUTHORIZATION = "Authorization"
const WWW_AUTHENTICATE = "WWW-Authenticate"
const STATUS = "Status"
const LOCATION = "Location"
const HOST = "Host"
const COOKIE = "Cookie"
const USER_AGENT = "User-Agent"
const ACCEPT = "Accept"
const ACCEPT_LANGUAGE = "Accept-Language"
const ACCEPT_ENCODING = "Accept-Encoding"
const UPGRADE_INSECURE_REQUESTS = "Upgrade-Insecure-Requests"
const SERVER = "Server"
const X_FORWARDED_HOST = "X-Forwarded-Host"
const X_FORWARDED_FOR = "X-Forwarded-For"
const X_FORWARDED_PROTO = "X-Forwarded-Proto"
const X_FORWARDED_PORT = "X-Forwarded-Port"

const CONNECTION_TYPE_CLOSE = 1
const CONNECTION_TYPE_KEEP_ALIVE = 2
const CONNECTION_TYPE_UPGRADE = 3
const CONNECTION_TYPE_UNKNOWN = 4

type Headers struct {
	status  int
	headers map[string][]string
}

func NewHeaders() *Headers {
	return &Headers{
		status:  httpstatus.OK,
		headers: map[string][]string{},
	}
}

func (h *Headers) Status() int {
	return h.status
}

func (h *Headers) SetStatus(status int) {
	h.status = status
}

func (h *Headers) CopyTo(dest *Headers) {
	dest.status = h.status
	for name, values := range h.headers {
		newValues := arrayutil.CopyArray(values)
		dest.headers[name] = newValues
	}
}

func (h *Headers) Get(name string) string {
	values := h.headers[strings.ToLower(name)]
	if values == nil {
		return ""
	} else {
		return values[0]
	}
}

func (h *Headers) GetInt(name string) (int, exception.Exception) {
	val := h.Get(name)
	if val == "" {
		return -1, nil
	} else {
		ival, err := strconv.Atoi(val)
		if err != nil {
			return -1, exception.NewExceptionFromError(err)
		} else {
			return ival, nil
		}
	}
}

func (h *Headers) Set(name string, value string) {
	name = strings.ToLower(name)
	h.headers[name] = []string{value}
}

func (h *Headers) SetInt(name string, value int) {
	h.Set(name, strconv.Itoa(value))
}

func (h *Headers) Add(name string, value string) {
	name = strings.ToLower(name)
	values := h.headers[name]
	if values == nil {
		values = []string{value}
	} else {
		values = append(values, value)
	}
	h.headers[name] = values
}

func (h *Headers) HeaderNames() []string {
	names := []string{}
	for name := range h.headers {
		names = append(names, name)
	}
	return names
}

func (h *Headers) HeaderValues(name string) []string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Contains(name string) bool {
	_, exists := h.headers[strings.ToLower(name)]
	return exists
}

func (h *Headers) Remove(name string) {
	delete(h.headers, strings.ToLower(name))
}

func (h *Headers) ContentType() string {
	return h.Get(CONTENT_TYPE)
}

func (h *Headers) SetContentType(typ string) {
	h.Set(CONTENT_TYPE, typ)
}

func (h *Headers) ContentLength() int {
	val, err := h.GetInt(CONTENT_LENGTH)
	if err != nil {
		return -1
	} else {
		return val
	}
}

func (h *Headers) SetContentLength(length int) {
	h.Set(CONTENT_LENGTH, strconv.Itoa(length))
}

func (h *Headers) GetConnection() int {
	con := h.Get(CONNECTION)
	return GetConnectionType(con)
}

func (h *Headers) Clear() {
	h.headers = map[string][]string{}
	h.status = httpstatus.OK
}

func GetConnectionType(typ string) int {
	switch strings.ToLower(typ) {
	case "keep-alive":
		return CONNECTION_TYPE_KEEP_ALIVE
	case "close":
		return CONNECTION_TYPE_CLOSE
	case "upgrade":
		return CONNECTION_TYPE_UPGRADE
	default:
		return CONNECTION_TYPE_UNKNOWN
	}
}
