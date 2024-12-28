package h2_error_code

import (
	"bayserver-core/baykit/bayserver/bayserver"
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/util"
)

const NO_ERROR = 0x0

const PROTOCOL_ERROR = 0x1

const INTERNAL_ERROR = 0x2

const FLOW_CONTROL_ERROR = 0x3

const SETTINGS_TIMEOUT = 0x4

const STREAM_CLOSED = 0x5

const FRAME_SIZE_ERROR = 0x6

const REFUSED_STREAM = 0x7

const CANCEL = 0x8

const COMPRESSION_ERROR = 0x9

const CONNECT_ERROR = 0xa

const ENHANCE_YOUR_CALM = 0xb

const INADEQUATE_SECURITY = 0xc

const HTTP_1_1_REQUIRED = 0xd

var Msg util.Message = nil

type h2ErrorCode struct {
	util.Message
}

func Init() bcf.ParseException {
	if Msg != nil {
		return nil
	}

	var ex bcf.ParseException
	Msg, ex = bayserver.LoadMessage("resources/conf/h2_messages", util.DefaultLocale())
	return ex
}
