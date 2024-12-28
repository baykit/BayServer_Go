package impl

import "strconv"

type CommandBase struct {
	typ int
}

func NewCommandBase(typ int) *CommandBase {
	cmd := &CommandBase{
		typ: typ,
	}

	return cmd
}

func (c *CommandBase) String() string {
	return "Command(" + strconv.Itoa(c.typ) + ")"
}

func (c *CommandBase) Type() int {
	return c.typ
}
