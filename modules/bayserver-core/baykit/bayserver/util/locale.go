package util

import (
	"os"
	"strings"
)

type Locale struct {
	Language string
	Country  string
}

func NewLocale(language string, country string) *Locale {
	return &Locale{language, country}
}

func DefaultLocale() *Locale {
	var lang = os.Getenv("LANG")
	if lang != "" {
		language := lang[0:2]
		country := lang[3:5]
		return NewLocale(language, country)
	}
	return NewLocale("en", "US")
}

func ParseLocale(locale string) *Locale {

	parts := strings.Split(locale, "-")
	if len(parts) >= 2 {
		return NewLocale(parts[0], parts[1])

	} else {
		return NewLocale(parts[0], "")
	}
}
