package docker

const TROUBLE_METHOD_GUIDE = 1
const TROUBLE_METHOD_TEXT = 2
const TROUBLE_METHOD_REROUTE = 3

type TroubleCommand struct {
	Method int
	Target string
}

func NewTroubleCommand(method int, target string) *TroubleCommand {
	return &TroubleCommand{method, target}
}

type Trouble interface {
	Docker
	Find(status int) *TroubleCommand
}
