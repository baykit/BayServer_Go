package tour

import (
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
	"bayserver-core/baykit/bayserver/util/httputil"
)

type RemoteHostResolver func() string

func NewDefaultRemoteHostResolver(ip string) RemoteHostResolver {
	return func() string {
		if ip == "" {
			return ""
		} else {
			host, ioerr := httputil.ResolveHost(ip)
			if ioerr != nil {
				baylog.Debug("Cannot resolve host %s: %s", ip, ioerr.Error())
				return ""
			} else {
				return host
			}
		}
	}
}

type TourReq interface {
	Init(key int)

	Key() int

	Uri() string
	SetUri(uri string)
	Method() string
	SetMethod(method string)
	Protocol() string
	SetProtocol(protocol string)
	ReqHost() string
	ReqPort() int

	RemoteAddress() string
	SetRemoteAddress(string)
	RemotePort() int
	SetRemotePort(int)
	RemoteHost() string
	SetRemoteHostFunc(resolver RemoteHostResolver)

	ServerAddress() string
	SetServerAddress(string)
	ServerPort() int
	SetServerPort(int)
	ServerName() string
	SetServerName(string)

	Headers() *headers.Headers

	SetLimit(length int)
	QueryString() string
	SetQueryString(queryString string)
	ScriptName() string
	SetScriptName(name string)
	PathInfo() string
	SetPathInfo(info string)
	Charset() string
	SetCharset(charset string)
	RewrittenUri() string
	SetRewrittenUri(uri string)

	ParseAuthorization()
	ParseHostPort(defaultPort int)

	GetReqContentHandler() ReqContentHandler
	SetReqContentHandler(handler ReqContentHandler)
	Consumed(checkId int, length int, lis ContentConsumeListener)
	PostReqContent(checkId int, data []byte, start int, len int, lis ContentConsumeListener) (bool, exception2.HttpException)
	EndReqContent(checkId int) (exception.IOException, exception2.HttpException)
	RemoteUser() string
	RemotePass() string

	BytesPosted() int
	BytesLimit() int
	Abort() bool
}
