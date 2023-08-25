package template

func trunc(length any, s string) string {
	c, err := toInt(length)
	if err != nil {
		panic(err)
	}

	if c < 0 && len(s)+c > 0 {
		return s[len(s)+c:]
	}

	if c >= 0 && len(s) > c {
		return s[:c]
	}

	return s
}
