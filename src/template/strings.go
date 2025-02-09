package template

import "unicode/utf8"

func trunc(length any, s string) string {
	c, err := toInt(length)
	if err != nil {
		panic(err)
	}

	runes := []rune(s)
	if len(runes) <= c {
		return s
	}

	if c < 0 {
		return string(runes[len(runes)+c:])
	}

	return string(runes[0:c])
}

func truncE(length any, s string) string {
	c, err := toInt(length)
	if err != nil {
		panic(err)
	}

	truncateSymbol := "…"

	if c == 0 {
		return truncateSymbol
	}

	lenTruncateSymbol := utf8.RuneCountInString(truncateSymbol)
	if c < 0 {
		lenTruncateSymbol *= -1
	}
	c -= lenTruncateSymbol

	runes := []rune(s)
	if len(runes) <= c {
		return s
	}

	if c < 0 {
		return truncateSymbol + string(runes[len(runes)+c:])
	}

	return string(runes[0:c]) + truncateSymbol
}
