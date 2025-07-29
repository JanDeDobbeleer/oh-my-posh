package text

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/generics"
)

var builderPool *generics.Pool[*StringBuilder]

type StringBuilder strings.Builder

func init() {
	builderPool = generics.NewPool(func() *StringBuilder {
		return &StringBuilder{}
	})
}

// NewBuilder returns a StringBuilder from the pool
func NewBuilder() *StringBuilder {
	return builderPool.Get()
}

func (sb *StringBuilder) release() {
	if sb == nil {
		return
	}

	// Reset the StringBuilder to clear its content
	(*strings.Builder)(sb).Reset()
	builderPool.Put(sb)
}

// String returns the accumulated string.
func (sb *StringBuilder) String() string {
	text := (*strings.Builder)(sb).String()
	sb.release()
	return text
}

// Len returns the number of accumulated bytes; b.Len() == len(b.String()).
func (sb *StringBuilder) Len() int {
	return (*strings.Builder)(sb).Len()
}

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (sb *StringBuilder) Cap() int {
	return (*strings.Builder)(sb).Cap()
}

// Reset resets the Builder to be empty.
func (sb *StringBuilder) Reset() {
	(*strings.Builder)(sb).Reset()
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation. If n is negative, Grow panics.
func (sb *StringBuilder) Grow(n int) {
	(*strings.Builder)(sb).Grow(n)
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
func (sb *StringBuilder) WriteRune(r rune) {
	_, _ = (*strings.Builder)(sb).WriteRune(r)
}

// WriteString appends the contents of s to b's buffer.
func (sb *StringBuilder) WriteString(s string) {
	_, _ = (*strings.Builder)(sb).WriteString(s)
}
