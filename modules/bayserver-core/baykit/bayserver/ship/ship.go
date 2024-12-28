package ship

import (
	"bayserver-core/baykit/bayserver/common"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/rudder"
	"bayserver-core/baykit/bayserver/util"
	exception2 "bayserver-core/baykit/bayserver/util/exception"
	"fmt"
)

const SHIP_ID_NOCHECK int = -1

const INVALID_SHIP_ID int = 0

type Ship interface {
	util.Reusable
	fmt.Stringer
	ObjectId() int
	ShipId() int
	AgentId() int
	Rudder() rudder.Rudder
	Transporter() common.Transporter

	NotifyHandshakeDone(protocol string) (common.NextSocketAction, exception2.IOException)
	NotifyConnect() (common.NextSocketAction, exception2.IOException)
	NotifyRead(buf []byte) (common.NextSocketAction, exception2.IOException)
	NotifyEof() common.NextSocketAction
	NotifyError(e exception2.Exception)
	NotifyProtocolError(e exception.ProtocolException) (bool, exception2.IOException)
	NotifyClose()
	CheckTimeout(durationSec int) bool

	CheckShipId(checkId int)
	ResumeRead(checkId int)
	PostClose(checkId int)
}
