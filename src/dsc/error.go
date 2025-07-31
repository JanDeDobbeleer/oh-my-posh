package dsc

type Error struct {
	message string
}

func (e *Error) Error() string {
	return `{
    "error": "` + e.message + `"
}`
}

func newError(message string) *Error {
	return &Error{
		message: message,
	}
}
