package dsc

func Error(err error) string {
	if err == nil {
		return ""
	}

	return `{
    "error": "` + err.Error() + `"
}`
}
