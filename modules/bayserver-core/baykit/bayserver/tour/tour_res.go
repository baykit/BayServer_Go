package tour

import (
	exception2 "bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/util/exception"
	"bayserver-core/baykit/bayserver/util/headers"
)

type TourRes interface {
	Init()
	Charset() string
	SetCharset(charset string)
	Headers() *headers.Headers
	HeaderSent() bool

	SendHeaders(checkId int) exception.IOException
	SendResContent(checkId int, buf []byte, start int, length int) (bool, exception.IOException)
	EndResContent(checkId int) exception.IOException
	SendError(checkId int, status int, message string, err exception.Exception) exception.IOException
	SendHttpException(checkId int, hterr exception2.HttpException) exception.IOException
	SetConsumeListener(lis ContentConsumeListener)
	DetachConsumeListener()

	BytesPosted() int
	BytesLimit() int
}
