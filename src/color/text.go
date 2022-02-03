package color

import "unicode/utf8"

func measureText(text string) int {
	length := utf8.RuneCountInString(text)
	return length
}
