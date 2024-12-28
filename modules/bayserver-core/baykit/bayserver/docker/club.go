package docker

import (
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/tour"
)

type Club interface {
	Docker

	/**
	 * Get the file name part of club
	 * @return
	 */
	FileName() string

	/**
	 * Get the ext (file extension part) of club
	 * @return
	 */
	Extension() string

	/**
	 * Check if file name matches this club
	 * @param fname
	 * @return
	 */
	Matches(fname string) bool

	/**
	 * Get charset of club
	 * @return
	 */
	Charset() string

	/**
	 * Check if this club decodes PATH_INFO
	 * @return
	 */
	DecodePathInfo() bool

	/**
	 * Arrive
	 */
	Arrive(tur tour.Tour) exception.HttpException
}
