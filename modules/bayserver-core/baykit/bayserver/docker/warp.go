package docker

import (
	"bayserver-core/baykit/bayserver/ship"
)

type Warp interface {
	Host() string

	Port() int

	DestTown() string

	TimeoutSec() int

	Keep(ship ship.Ship)

	OnEndShip(ship ship.Ship)
}
