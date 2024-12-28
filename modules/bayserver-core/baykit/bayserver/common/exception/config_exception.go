package exception

import (
	"strconv"
)

type ConfigException interface {
	BayException

	GetFile() string
	GetLineNo() int
}

type ConfigExceptionImpl struct {
	BayExceptionImpl

	file   string
	lineNo int
}

func (c *ConfigExceptionImpl) GetFile() string {
	return c.file
}

func (c *ConfigExceptionImpl) GetLineNo() int {
	return c.lineNo
}

func NewConfigException(file string, lineNo int, format string, args ...interface{}) ConfigException {
	ex := &ConfigExceptionImpl{}
	ex.ConstructException(4, nil, format, args...)
	ex.file = file
	ex.lineNo = lineNo
	ex.Message = CreatePositionMessage(ex.Message, file, lineNo)

	// interface check
	var _ BayException = ex
	var _ ConfigException = ex

	return ex
}

func CreatePositionMessage(msg string, file string, line int) string {
	return msg + " (at " + file + ":" + strconv.Itoa(line) + ")"
}
