package impl

import (
	"bayserver-core/baykit/bayserver/bcf"
	"bayserver-core/baykit/bayserver/common/baymessage"
	"bayserver-core/baykit/bayserver/common/exception"
	"bayserver-core/baykit/bayserver/symbol"
	"bayserver-core/baykit/bayserver/util/arrayutil"
	"bayserver-core/baykit/bayserver/util/baylog"
	"bayserver-core/baykit/bayserver/util/strutil"
	"bufio"
	"embed"
	"io"
	"io/fs"
	"os"
	"strings"
	"unicode"
)

type lineInfo struct {
	lineObj interface{} // *BcfObject
	indent  int
}

type BcfParserImpl struct {
	bcf.BcfParser
	lineNo       int
	prevLineInfo *lineInfo
	fileName     string
	indentMap    []int
	reader       *bufio.Reader
}

func NewBcfParser() BcfParserImpl {
	return BcfParserImpl{
		lineNo:       -1,
		prevLineInfo: nil,
		fileName:     "",
		indentMap:    []int{},
		reader:       nil,
	}
}

func newLineInfo(lineObj interface{}, indent int) *lineInfo {
	return &lineInfo{lineObj, indent}
}

func (p *BcfParserImpl) pushIndent(spCount int) {
	p.indentMap = append(p.indentMap, spCount)
}

func (p *BcfParserImpl) popIndent() {
	p.indentMap = p.indentMap[:len(p.indentMap)-1]
}

func (p *BcfParserImpl) getIndent(spCount int) (int, bcf.ParseException) {
	if len(p.indentMap) == 0 {
		p.pushIndent(spCount)

	} else if spCount > p.indentMap[len(p.indentMap)-1] {
		p.pushIndent(spCount)
	}

	indent := arrayutil.IndexOf(p.indentMap, spCount)
	if indent == -1 {
		return -1, bcf.NewParseException(
			p.fileName,
			p.lineNo,
			baymessage.Get(symbol.PAS_INVALID_INDENT))
	}
	return indent, nil
}

func (p *BcfParserImpl) Parse(fileName string) (*bcf.BcfDocument, bcf.ParseException) {

	file, err := os.Open(fileName)

	if err != nil {
		return nil, bcf.NewParseException(fileName, 0, err.Error())
	}
	defer file.Close()

	return p.ParseFile(fileName, file)
}

func (p *BcfParserImpl) ParseResource(res *embed.FS, fileName string) (*bcf.BcfDocument, bcf.ParseException) {

	file, err := res.Open(fileName)

	if err != nil {
		return nil, bcf.NewParseException(fileName, 0, err.Error())
	}
	defer file.Close()

	return p.ParseFile(fileName, file)
}

func (p *BcfParserImpl) ParseFile(fileName string, file fs.File) (*bcf.BcfDocument, bcf.ParseException) {
	doc := bcf.NewBcfDocument()

	p.fileName = fileName
	p.lineNo = 0
	p.reader = bufio.NewReader(file)

	_, ex := p.parseSameLevel(&doc.ContentList, 0)
	if ex != nil {
		baylog.ErrorE(ex, "")
		return nil, bcf.NewParseException(p.fileName, p.lineNo, ex.Error())
	}
	return &doc, nil
}

func (p *BcfParserImpl) parseSameLevel(curList *[]interface{}, indent int) (*lineInfo, bcf.ParseException) {
	objectExistsInSameLevel := false
	for {
		var lineInfo *lineInfo = nil
		if p.prevLineInfo != nil {
			lineInfo = p.prevLineInfo
			p.prevLineInfo = nil

		} else {
			line, err := p.reader.ReadString('\n')
			p.lineNo++
			if err != nil {
				if err != io.EOF {
					bayex := exception.NewBayExceptionFromError(err)
					baylog.ErrorE(bayex, "")
					return nil, bcf.NewParseException(
						p.fileName,
						p.lineNo,
						"%s",
						err.Error())

				}
			}
			if line == "" {
				break
			}

			if strutil.StartsWith(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
				// comment or empty line
				continue
			}

			var ex bcf.ParseException
			lineInfo, ex = p.parseLine(p.lineNo, line)
			if ex != nil {
				return nil, ex
			}
		}

		if lineInfo == nil {
			// comment or empty
			continue

		} else if lineInfo.indent > indent {
			// lower level
			return nil, bcf.NewParseException(
				p.fileName,
				p.lineNo,
				baymessage.Get(symbol.PAS_INVALID_INDENT))

		} else if lineInfo.indent < indent {
			// upper level
			p.prevLineInfo = lineInfo
			if objectExistsInSameLevel {
				p.popIndent()
			}
			return lineInfo, nil

		} else {
			objectExistsInSameLevel = true
			// same level
			if elm, ok := lineInfo.lineObj.(*bcf.BcfElement); ok {
				*curList = append(*curList, lineInfo.lineObj)

				lastLineInfo, err := p.parseSameLevel(&elm.ContentList, lineInfo.indent+1)
				if err != nil {
					return nil, err
				}
				if lastLineInfo == nil {
					// EOF
					p.popIndent()
					return nil, nil

				} else {
					// Same level
					continue
				}

			} else {
				*curList = append(*curList, lineInfo.lineObj)
				continue
			}
		}
	}
	p.popIndent()
	return nil, nil
}

func (p *BcfParserImpl) parseLine(lineNo int, line string) (*lineInfo, bcf.ParseException) {
	var spCount = -1
	for spCount = 0; spCount < len(line); spCount++ {
		c := line[spCount]
		if !unicode.IsSpace(rune(c)) {
			break
		}

		if c != ' ' {
			return nil, bcf.NewParseException(
				p.fileName,
				p.lineNo,
				baymessage.Get(symbol.PAS_INVALID_WHITESPACE))
		}
	}
	indent, ex := p.getIndent(spCount)
	if ex != nil {
		return nil, ex
	}

	line = line[spCount:]
	line = strings.TrimSpace(line)

	if strutil.StartsWith(line, "[") {
		closePos := strings.Index(line, "]")
		if closePos == -1 {
			return nil, bcf.NewParseException(
				p.fileName,
				p.lineNo,
				baymessage.Get(symbol.PAS_BRACE_NOT_CLOSED))
		}
		if !strutil.EndsWith(line, "]") {
			return nil, bcf.NewParseException(
				p.fileName,
				p.lineNo,
				baymessage.Get(symbol.PAS_INVALID_LINE))
		}
		keyVal := p.parseKeyVal(line[1:closePos], lineNo)
		return newLineInfo(
			bcf.NewBcfElement(keyVal.Key, keyVal.Value, p.fileName, p.lineNo),
			indent), nil

	} else {
		return &lineInfo{p.parseKeyVal(line, lineNo), indent}, nil
	}

}

func (p *BcfParserImpl) parseKeyVal(line string, lineNo int) *bcf.BcfKeyVal {
	spPos := strings.Index(line, " ")
	var key string
	var val string
	if spPos == -1 {
		key = line
		val = ""

	} else {
		key = line[0:spPos]
		val = strings.TrimSpace(line[spPos:])
	}
	return bcf.NewBcfKeyVal(key, val, p.fileName, lineNo)
}
