package docker

type Reroute interface {
	Docker

	Reroute(twn Town, uri string) string
}
