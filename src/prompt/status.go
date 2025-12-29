package prompt

func (e *Engine) Status() string {
	e.writePrimaryPrompt(false)
	return e.string()
}
