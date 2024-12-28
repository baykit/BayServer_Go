package protocol

import "bayserver-core/baykit/bayserver/util"

/**
 * base class for handling commands
 * (Uses visitor pattern)
 */

type CommandHandler interface {
	util.Reusable
}
