package segments

import "strings"

// extractRequiredVersion scans HCL source for a top-level terraform block and
// returns the value of its required_version attribute. It is not an HCL
// parser: it only tracks the structure needed to locate that one attribute -
// comments, strings (including template interpolation), heredocs and block
// nesting. Whenever the source doesn't match what it expects, it reports
// false and the segment falls back to the state file, the same path a full
// parser failure took before.
func extractRequiredVersion(src string) (string, bool) {
	s := &hclScanner{src: src, terraformDepth: -1}

	for !s.eof() {
		s.skipTrivia()
		if s.eof() {
			break
		}

		switch c := s.src[s.pos]; {
		case c == '"':
			s.skipString()
		case c == '<' && s.peek() == '<':
			s.skipHeredoc()
		case c == '{':
			s.depth++
			s.pos++
		case c == '}':
			s.depth--
			if s.depth < s.terraformDepth {
				s.terraformDepth = -1
			}
			s.pos++
		case isHCLIdentStart(c):
			ident := s.readIdent()
			if version, ok := s.handleIdent(ident); ok {
				return version, true
			}
		default:
			s.pos++
		}
	}

	return "", false
}

type hclScanner struct {
	src            string
	pos            int
	depth          int
	terraformDepth int
}

// handleIdent inspects an identifier just read at the current depth. A
// top-level terraform identifier followed by { opens the block we're after;
// required_version directly inside it yields the result.
func (s *hclScanner) handleIdent(ident string) (string, bool) {
	if s.depth == 0 && ident == "terraform" {
		s.skipTrivia()
		if !s.eof() && s.src[s.pos] == '{' {
			s.depth++
			s.terraformDepth = s.depth
			s.pos++
		}
		return "", false
	}

	if ident != "required_version" || s.terraformDepth == -1 || s.depth != s.terraformDepth {
		return "", false
	}

	s.skipTrivia()
	if s.eof() || s.src[s.pos] != '=' {
		return "", false
	}

	s.pos++
	s.skipTrivia()
	if s.eof() || s.src[s.pos] != '"' {
		return "", false
	}

	return s.readString()
}

func (s *hclScanner) eof() bool {
	return s.pos >= len(s.src)
}

func (s *hclScanner) peek() byte {
	if s.pos+1 >= len(s.src) {
		return 0
	}

	return s.src[s.pos+1]
}

// skipTrivia advances past whitespace and comments (#, // and /* */).
func (s *hclScanner) skipTrivia() {
	for !s.eof() {
		switch c := s.src[s.pos]; {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			s.pos++
		case c == '#' || (c == '/' && s.peek() == '/'):
			for !s.eof() && s.src[s.pos] != '\n' {
				s.pos++
			}
		case c == '/' && s.peek() == '*':
			s.pos += 2
			for !s.eof() && (s.src[s.pos] != '*' || s.peek() != '/') {
				s.pos++
			}
			s.pos += 2
		default:
			return
		}
	}
}

// skipString advances past a quoted string, honoring escapes and ${ }
// template interpolation, which may itself contain strings and braces.
func (s *hclScanner) skipString() {
	_, _ = s.readString()
}

// readString consumes a quoted string starting at the opening quote and
// returns its content with common escapes resolved. Interpolation is copied
// verbatim; required_version must be a static string, so a value containing
// ${ } simply won't match a version constraint downstream.
func (s *hclScanner) readString() (string, bool) {
	var sb strings.Builder
	s.pos++ // opening quote

	for !s.eof() {
		switch c := s.src[s.pos]; {
		case c == '"':
			s.pos++
			return sb.String(), true
		case c == '\\':
			escaped, ok := hclEscape(s.peek())
			if ok {
				sb.WriteByte(escaped)
			}

			if !ok {
				sb.WriteByte(c)
				sb.WriteByte(s.peek())
			}

			s.pos += 2
		case c == '$' && s.peek() == '{':
			start := s.pos
			s.skipInterpolation()
			sb.WriteString(s.src[start:s.pos])
		default:
			sb.WriteByte(c)
			s.pos++
		}
	}

	return "", false
}

// skipInterpolation advances past a ${ } template expression, tracking brace
// nesting and skipping over nested strings so their braces don't miscount.
func (s *hclScanner) skipInterpolation() {
	s.pos += 2
	braces := 1

	for !s.eof() && braces > 0 {
		switch s.src[s.pos] {
		case '{':
			braces++
			s.pos++
		case '}':
			braces--
			s.pos++
		case '"':
			s.skipString()
		default:
			s.pos++
		}
	}
}

// skipHeredoc advances past a << or <<- heredoc, which runs until a line
// containing only the delimiter label.
func (s *hclScanner) skipHeredoc() {
	s.pos += 2
	if !s.eof() && s.src[s.pos] == '-' {
		s.pos++
	}

	start := s.pos
	for !s.eof() && s.src[s.pos] != '\n' {
		s.pos++
	}

	label := strings.TrimSpace(s.src[start:s.pos])
	if label == "" {
		return
	}

	for !s.eof() {
		s.pos++ // the newline
		lineStart := s.pos
		for !s.eof() && s.src[s.pos] != '\n' {
			s.pos++
		}

		if strings.TrimSpace(s.src[lineStart:s.pos]) == label {
			return
		}
	}
}

func (s *hclScanner) readIdent() string {
	start := s.pos
	for !s.eof() && isHCLIdentPart(s.src[s.pos]) {
		s.pos++
	}

	return s.src[start:s.pos]
}

func hclEscape(c byte) (byte, bool) {
	switch c {
	case 'n':
		return '\n', true
	case 't':
		return '\t', true
	case 'r':
		return '\r', true
	case '"':
		return '"', true
	case '\\':
		return '\\', true
	default:
		return 0, false
	}
}

func isHCLIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isHCLIdentPart(c byte) bool {
	return isHCLIdentStart(c) || c == '-' || (c >= '0' && c <= '9')
}
