package progress

import (
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

func NewReader(reader io.Reader, total int64, program *tea.Program) *Reader {
	return &Reader{
		Reader:  reader,
		program: program,
		total:   total,
	}
}

type Reader struct {
	io.Reader

	program *tea.Program
	total   int64
	current int64
}

func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.current += int64(n)
	percent := float64(r.current) / float64(r.total)

	if r.program != nil {
		r.program.Send(Message(percent))
	}

	return n, err
}
