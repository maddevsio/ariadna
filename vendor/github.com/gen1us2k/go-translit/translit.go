package translit

import (
	"regexp"
)

// Translit transliterates english string to cyrillic analog
func Translit(s string) string {
	cyrillicString := s
	for k, v := range enTranslations {
		r, _ := regexp.Compile(k)
		cyrillicString = r.ReplaceAllString(cyrillicString, v)
	}
	return cyrillicString
}
