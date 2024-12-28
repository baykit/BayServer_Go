package impl

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/util"
)

var msg util.Message

func Init() exception.BayException {
	var err exception.BayException
	msg, err = bayserver.LoadMessage("resources/conf/messages", util.DefaultLocale())

	baymessage.Get = func(key string, args ...interface{}) string {
		return msg.Get(key, args...)
	}

	return err
}
