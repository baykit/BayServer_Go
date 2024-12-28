package cgiutil

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/tour"
	"bayserver-core/baykit/bayserver/util/headers"
	"os"
	path2 "path"
	"strconv"
	"strings"
)

const REQUEST_METHOD = "REQUEST_METHOD"
const REQUEST_URI = "REQUEST_URI"
const SERVER_PROTOCOL = "SERVER_PROTOCOL"
const GATEWAY_INTERFACE = "GATEWAY_INTERFACE"
const SERVER_NAME = "SERVER_NAME"
const SERVER_PORT = "SERVER_PORT"
const QUERY_STRING = "QUERY_STRING"
const SCRIPT_NAME = "SCRIPT_NAME"
const SCRIPT_FILENAME = "SCRIPT_FILENAME"
const PATH_TRANSLATED = "PATH_TRANSLATED"
const PATH_INFO = "PATH_INFO"
const CONTENT_TYPE = "CONTENT_TYPE"
const CONTENT_LENGTH = "CONTENT_LENGTH"
const REMOTE_ADDR = "REMOTE_ADDR"
const REMOTE_PORT = "REMOTE_PORT"
const REMOTE_USER = "REMOTE_USER"
const HTTP_ACCEPT = "HTTP_ACCEPT"
const HTTP_COOKIE = "HTTP_COOKIE"
const HTTP_HOST = "HTTP_HOST"
const HTTP_USER_AGENT = "HTTP_USER_AGENT"
const HTTP_ACCEPT_ENCODING = "HTTP_ACCEPT_ENCODING"
const HTTP_ACCEPT_LANGUAGE = "HTTP_ACCEPT_LANGUAGE"
const HTTP_CONNECTION = "HTTP_CONNECTION"
const HTTP_UPGRADE_INSECURE_REQUESTS = "HTTP_UPGRADE_INSECURE_REQUESTS"
const HTTPS = "HTTPS"
const PATH = "PATH"
const SERVER_SIGNATURE = "SERVER_SIGNATURE"
const SERVER_SOFTWARE = "SERVER_SOFTWARE"
const SERVER_ADDR = "SERVER_ADDR"
const DOCUMENT_ROOT = "DOCUMENT_ROOT"
const REQUEST_SCHEME = "REQUEST_SCHEME"
const CONTEXT_PREFIX = "CONTEXT_PREFIX"
const CONTEXT_DOCUMENT_ROOT = "CONTEXT_DOCUMENT_ROOT"
const SERVER_ADMIN = "SERVER_ADMIN"
const REQUEST_TIME_FLOAT = "REQUEST_TIME_FLOAT"
const REQUEST_TIME = "REQUEST_TIME"
const UNIQUE_ID = "UNIQUE_ID"

type CallBack func(name string, value string)

func GetEnv(path string, docRoot string, scriptBase string, tur tour.Tour) map[string]string {

	m := make(map[string]string)
	GetEnv2(path, docRoot, scriptBase, tur, func(name string, value string) { m[name] = value })
	return m
}

func GetEnv2(path string, docRoot string, scriptBase string, tur tour.Tour, cb CallBack) {

	reqHeaders := tur.Req().Headers()

	ctype := reqHeaders.ContentType()
	if ctype != "" {
		pos := strings.Index(ctype, "charset=")
		if pos >= 0 {
			tur.Req().SetCharset(strings.TrimSpace(ctype[pos+8:]))
		}
	}

	addEnv(cb, REQUEST_METHOD, tur.Req().Method())
	addEnv(cb, REQUEST_URI, tur.Req().Uri())
	addEnv(cb, SERVER_PROTOCOL, tur.Req().Protocol())
	addEnv(cb, GATEWAY_INTERFACE, "CGI/1.1")

	addEnv(cb, SERVER_NAME, tur.Req().ReqHost())
	addEnv(cb, SERVER_ADDR, tur.Req().ServerAddress())
	if tur.Req().ReqPort() >= 0 {
		addEnv(cb, SERVER_PORT, strconv.Itoa(tur.Req().ReqPort()))
	}
	addEnv(cb, SERVER_SOFTWARE, bayserver.SoftwareName())

	addEnv(cb, CONTEXT_DOCUMENT_ROOT, docRoot)

	for _, name := range tur.Req().Headers().HeaderNames() {
		newVal := ""
		for _, value := range tur.Req().Headers().HeaderValues(name) {
			if newVal == "" {
				newVal = value
			} else {
				newVal = newVal + "; " + value
			}
		}

		name = strings.Replace(strings.ToUpper(name), "-", "_", -1)
		if strings.HasPrefix(name, "X_FORWARDED_") {
			addEnv(cb, name, newVal)

		} else {
			switch name {
			case CONTENT_TYPE, CONTENT_LENGTH:
				addEnv(cb, name, newVal)

			default:
				addEnv(cb, "HTTP_"+name, newVal)
			}
		}

		addEnv(cb, REMOTE_ADDR, tur.Req().RemoteAddress())
		addEnv(cb, REMOTE_PORT, strconv.Itoa(tur.Req().RemotePort()))

		if tur.Secure() {
			addEnv(cb, REQUEST_SCHEME, "https")
		} else {
			addEnv(cb, REQUEST_SCHEME, "http")
		}

		tmpSecure := tur.Secure()
		fproto := tur.Req().Headers().Get(headers.X_FORWARDED_PROTO)
		if fproto != "" {
			tmpSecure = strings.ToLower(fproto) == "https"
		}
		if tmpSecure {
			addEnv(cb, HTTPS, "on")
		}

		addEnv(cb, QUERY_STRING, tur.Req().QueryString())
		addEnv(cb, SCRIPT_NAME, tur.Req().ScriptName())

		if tur.Req().PathInfo() == "" {
			addEnv(cb, PATH_INFO, "")

		} else {
			addEnv(cb, PATH_INFO, tur.Req().PathInfo())
			pathTranslated := path2.Join(docRoot, tur.Req().PathInfo())
			addEnv(cb, PATH_TRANSLATED, pathTranslated)

		}

		if !strings.HasSuffix(scriptBase, "/") {
			scriptBase += "/"
		}
		addEnv(cb, SCRIPT_FILENAME, scriptBase+tur.Req().ScriptName()[len(path):])
		addEnv(cb, PATH, os.Getenv("PATH"))
	}

}

func addEnv(cb CallBack, key string, value string) {
	cb(key, value)
}
