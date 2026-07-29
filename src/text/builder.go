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

func (sb *StringBuilder) String() string {
	text := (*strings.Builder)(sb).String()
	sb.release()
	return text
}

func (sb *StringBuilder) Len() int {
	return (*strings.Builder)(sb).Len()
}

func (sb *StringBuilder) Cap() int {
	return (*strings.Builder)(sb).Cap()
}

func (sb *StringBuilder) Reset() {
	(*strings.Builder)(sb).Reset()
}

// If n is negative, Grow panics.
func (sb *StringBuilder) Grow(n int) {
	(*strings.Builder)(sb).Grow(n)
}

func (sb *StringBuilder) WriteRune(r rune) {
	_, _ = (*strings.Builder)(sb).WriteRune(r)
}

func (sb *StringBuilder) WriteString(s string) {
	_, _ = (*strings.Builder)(sb).WriteString(s)
}
