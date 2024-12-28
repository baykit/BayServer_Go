package impl

import (
	"bayserver-core/baykit/bayserver/protocol"
)

type CommandUnpackerImpl struct {
	protocol.CommandUnpacker
}

/****************************************/
/* Implements Reusable                  */
/****************************************/

func (cu *CommandUnpackerImpl) Reset() {

}
