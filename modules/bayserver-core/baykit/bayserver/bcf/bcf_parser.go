package bcf

import (
	"bayserver-core/baykit/bayserver/common/exception"
	"os"
)

type BcfParser interface {
	Parse(fileName string) (*BcfDocument, exception.BayException)
	ParseFile(fileName string, file *os.File) (*BcfDocument, ParseException)
}
