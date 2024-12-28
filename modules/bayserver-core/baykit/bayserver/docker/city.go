package docker

import (
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/tour"
)

type City interface {
	Docker

	/**
	 * City name (host name)
	 * @return city name
	 */
	Name() string

	/**
	 * All clubs (not included in town) in this city
	 * @return arrayutil of club
	 */
	Clubs() []Club

	/**
	 * All towns in this city
	 * @return arrayutil of town
	 */
	Towns() []Town

	/**
	 * Enter city
	 * @param tour
	 */
	Enter(tur tour.Tour) exception.HttpException

	/**
	 * Get trouble base
	 * @return
	 */
	GetTrouble() Trouble

	/**
	 * Logging
	 * @param tour
	 */
	Log(tur tour.Tour)
}
