package docker

import (
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/tour"
)

var MATCH_TYPE_MATCHED = 1
var MATCH_TYPE_NOT_MATCHED = 2
var MATCH_TYPE_CLOSE = 3

type Town interface {
	Docker

	/**
	 * Get the name (path) of this town
	 * The name ends with "/"
	 */
	Name() string

	/**
	 * Get city
	 */
	City() City

	/**
	 * Get the physical location of this town
	 */
	Location() string

	/**
	 * Get index file
	 */
	WelcomeFile() string

	/**
	 * All clubs in this town
	 * @return club list
	 */
	Clubs() []Club

	/**
	 * Get rerouted uri
	 * @return reroute list
	 */
	Reroute(uri string) string

	Matches(uri string) int

	CheckAdmitted(tur tour.Tour) exception.HttpException
}
