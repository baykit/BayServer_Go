package charutil

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"strings"
)

func Lower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		c += 'a' - 'A'
	}
	return c
}

func Upper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}
	return c
}

func GetEncoding(charset string) encoding.Encoding {
	switch strings.ToLower(charset) {
	case "utf-8":
		return unicode.UTF8
	case "utf-16":
		return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
	case "utf-16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	case "shift-jis":
		return japanese.ShiftJIS
	case "euc-jp":
		return japanese.EUCJP
	case "iso-2022-jp":
		return japanese.ISO2022JP
	case "euc-kr":
		return korean.EUCKR
	case "gbk":
		return simplifiedchinese.GBK
	case "gb18030":
		return simplifiedchinese.GB18030
	case "big5":
		return traditionalchinese.Big5
	case "iso-8859-1":
		return charmap.ISO8859_1
	default:
		return nil
	}
}
