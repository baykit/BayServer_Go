package logitems

import "bayserver-core/baykit/bayserver/tour"

type LogItem interface {

	/**
	 * initialize
	 */

	Init(param string)

	/**
	 * Print log
	 */

	GetItem(tour tour.Tour) string
}
