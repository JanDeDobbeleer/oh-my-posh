package terminal

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/regex"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
	"github.com/mattn/go-runewidth"
)

func init() {
	runewidth.DefaultCondition.EastAsianWidth = false
}

type style struct {
	AnchorStart string
	AnchorEnd   string
	Start       string
	End         string
}

var (
	knownStyles = []*style{
		{AnchorStart: `<b>`, AnchorEnd: `</b>`, Start: "\x1b[1m", End: "\x1b[22m"},
		{AnchorStart: `<u>`, AnchorEnd: `</u>`, Start: "\x1b[4m", End: "\x1b[24m"},
		{AnchorStart: `<o>`, AnchorEnd: `</o>`, Start: "\x1b[53m", End: "\x1b[55m"},
		{AnchorStart: `<i>`, AnchorEnd: `</i>`, Start: "\x1b[3m", End: "\x1b[23m"},
		{AnchorStart: `<s>`, AnchorEnd: `</s>`, Start: "\x1b[9m", End: "\x1b[29m"},
		{AnchorStart: `<d>`, AnchorEnd: `</d>`, Start: "\x1b[2m", End: "\x1b[22m"},
		{AnchorStart: `<f>`, AnchorEnd: `</f>`, Start: "\x1b[5m", End: "\x1b[25m"},
		{AnchorStart: `<r>`, AnchorEnd: `</r>`, Start: "\x1b[7m", End: "\x1b[27m"},
	}

	resetStyle      = &style{AnchorStart: "RESET", AnchorEnd: `</>`, End: "\x1b[0m"}
	backgroundStyle = &style{AnchorStart: "BACKGROUND", AnchorEnd: `</>`, End: "\x1b[49m"}

	BackgroundColor color.Ansi
	CurrentColors   *color.Set
	ParentColors    []*color.Set
	Colors          color.String

	Plain       bool
	Interactive bool

	builder strings.Builder
	length  int

	foregroundColor color.Ansi
	backgroundColor color.Ansi
	currentColor    color.History
	textLen         int

	// bgGradientCells/fgGradientCells hold one ready-to-print ANSI code per visible
	// cell of the segment being written, populated by color.GradientCells when the
	// corresponding channel is a gradient. cellIndex is the shared cursor into both
	// slices, advanced once per visible rune regardless of which channel(s) stamp.
	// See stampGradient/writeVisibleRune.
	bgGradientCells []color.Ansi
	fgGradientCells []color.Ansi
	cellIndex       int

	// gradientRenderCells is the segment currently being written's visible cell count,
	// set once cells is known (see Write). collapseGradientLast reads it so a
	// dark-gradient/light-gradient color override edge mid-body matches the same shade
	// GradientCells rendered the segment's actual last cell as (see GradientLastForCells).
	// Zero (its reset value) falls back to GradientLast's gentlest single-step shade.
	gradientRenderCells int

	isTransparent bool
	isInvisible   bool
	isHyperlink   bool

	Shell   string
	Program string

	progressTerminals []string

	formats *shell.Formats

	// escapePrefix/escapeSuffix are formats.Escape ("...%s...") split around its
	// single %s placeholder, precomputed once so writeEscapedAnsiString can
	// concatenate via the builder instead of allocating through fmt.Sprintf.
	escapePrefix string
	escapeSuffix string
)

const (
	AnchorRegex = `^(?P<ANCHOR><(?P<FG>[^,<>]+)?,?(?P<BG>[^<>]+)?>)`

	// colorisePrefix/coloriseSuffix and transparentStartPrefix/transparentStartSuffix
	// are the fixed parts of the colorise ("\x1b[%sm") and transparentStart
	// ("\x1b[0m\x1b[%s;49m\x1b[7m") formats, split around their single %s
	// placeholder so callers can write them directly via the builder instead
	// of allocating through fmt.Sprintf.
	colorisePrefix = "\x1b["
	coloriseSuffix = "m"

	transparentStartPrefix = "\x1b[0m\x1b["
	transparentStartSuffix = ";49m\x1b[7m"

	transparentEnd = "\x1b[27m"
	backgroundEnd  = "\x1b[49m"

	AnsiRegex = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	OSC99 = "osc99"
	OSC7  = "osc7"
	OSC51 = "osc51"

	ANCHOR = "ANCHOR"
	BG     = "BG"
	FG     = "FG"

	hyperLinkStart   = "<LINK>"
	hyperLinkEnd     = "</LINK>"
	hyperLinkText    = "<TEXT>"
	hyperLinkTextEnd = "</TEXT>"

	empty = "<>"

	startProgress = "\x1b]9;4;3;0\x07"
	setProgress   = "\x1b]9;4;4;%d\x07"
	endProgress   = "\x1b]9;4;0;0\x07"

	WindowsTerminal = "Windows Terminal"
	Warp            = "WarpTerminal"
	ITerm           = "iTerm.app"
	AppleTerminal   = "Apple_Terminal"
	Unknown         = "Unknown"
)

// anchorMatch describes a single `<...>` anchor token found while scanning
// segment text, mirroring the named groups of AnchorRegex without allocating
// a map or invoking the regexp engine.
type anchorMatch struct {
	Anchor string
	FG     string
	BG     string
	ok     bool
}

// scanAnchor looks for an AnchorRegex-shaped token at the start of txt, i.e.
// `<` + zero or more non `<>` characters + `>`, with the inner text optionally
// split on the first comma into FG (before) and BG (after, which may itself
// contain further commas). It operates on the zero-copy slice txt[i:] and
// performs no allocations.
func scanAnchor(txt string) anchorMatch {
	if len(txt) == 0 || txt[0] != '<' {
		return anchorMatch{}
	}

	end := strings.IndexByte(txt, '>')
	if end < 0 {
		return anchorMatch{}
	}

	inner := txt[1:end]
	if strings.IndexByte(inner, '<') >= 0 {
		return anchorMatch{}
	}

	fg := inner
	bg := ""

	if before, after, ok := strings.Cut(inner, ","); ok {
		fg = before
		bg = after
	}

	return anchorMatch{
		Anchor: txt[:end+1],
		FG:     fg,
		BG:     bg,
		ok:     true,
	}
}

func Init(sh string) {
	Shell = sh
	Program = getTerminalName()

	log.Debug("terminal program:", Program)
	log.Debug("terminal shell:", Shell)

	color.TrueColor = Program != AppleTerminal

	progressTerminals = []string{WindowsTerminal}
	formats = shell.GetFormats(Shell)

	escapePrefix, escapeSuffix = "", ""
	if before, after, found := strings.Cut(formats.Escape, "%s"); found {
		// formats.Escape is a fmt.Sprintf format string, so any literal "%"
		// in the surrounding text is escaped as "%%" (e.g. zsh's "%%{%s%%}").
		// Unescape it now since we no longer route through fmt.Sprintf.
		escapePrefix = strings.ReplaceAll(before, "%%", "%")
		escapeSuffix = strings.ReplaceAll(after, "%%", "%")
	}
}

func getTerminalName() string {
	Program = os.Getenv("TERM_PROGRAM")
	if len(Program) != 0 {
		return Program
	}

	wtSession := os.Getenv("WT_SESSION")
	if len(wtSession) != 0 {
		return WindowsTerminal
	}

	return Unknown
}

func SetColors(background, foreground color.Ansi) {
	CurrentColors = &color.Set{
		Background: background,
		Foreground: foreground,
	}
}

// SetParentColors pushes the completed segment's colors onto the parent
// stack; the most recent entry (nearest ancestor) sits at the tail. Cleared
// per block by String() - see resolveParentColor in color/keywords.go for
// the matching tail-to-head walk.
func SetParentColors(background, foreground color.Ansi) {
	ParentColors = append(ParentColors, &color.Set{
		Background: background,
		Foreground: foreground,
	})
}

func ChangeLine(numberOfLines int) string {
	if Plain {
		return ""
	}

	position := "B"

	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}

	return fmt.Sprintf(formats.Linechange, numberOfLines, position)
}

func Pwd(pwdType, userName, hostName, pwd string) string {
	if Plain {
		return ""
	}

	switch pwdType {
	case OSC7:
		return fmt.Sprintf(formats.Osc7, hostName, pwd)
	case OSC51:
		return fmt.Sprintf(formats.Osc51, userName, hostName, pwd)
	case OSC99:
		fallthrough
	default:
		return fmt.Sprintf(formats.Osc99, pwd)
	}
}

func ClearAfter() string {
	if Plain {
		return ""
	}

	return formats.ClearLine + formats.ClearBelow
}

func FormatTitle(title string) string {
	switch Shell {
	// These shells don't support setting the console title.
	case shell.ELVISH, shell.XONSH:
		return ""
	case shell.BASH, shell.ZSH, shell.YASH:
		title = trimAnsi(title)

		sb := text.NewBuilder()

		// We have to do this to prevent the shell from misidentifying escape sequences.
		for _, char := range title {
			escaped, shouldEscape := formats.EscapeSequences[char]
			if shouldEscape {
				sb.WriteString(escaped)
				continue
			}

			sb.WriteRune(char)
		}

		return fmt.Sprintf(formats.Title, sb.String())
	default:
		return fmt.Sprintf(formats.Title, trimAnsi(title))
	}
}

func EscapeText(txt string) string {
	return fmt.Sprintf(formats.Escape, txt)
}

func SaveCursorPosition() string {
	return formats.SaveCursorPosition
}

func RestoreCursorPosition() string {
	return formats.RestoreCursorPosition
}

func PromptStart() string {
	return fmt.Sprintf(formats.Escape, "\x1b]133;A\007")
}

func CommandStart() string {
	return fmt.Sprintf(formats.Escape, "\x1b]133;B\007")
}

func CommandFinished(code int, ignore bool) string {
	if ignore {
		return fmt.Sprintf(formats.Escape, "\x1b]133;D\007")
	}

	mark := fmt.Sprintf("\x1b]133;D;%d\007", code)

	return fmt.Sprintf(formats.Escape, mark)
}

func LineBreak() string {
	cr := fmt.Sprintf(formats.Left, 1000)
	lf := fmt.Sprintf(formats.Linechange, 1, "B")
	return cr + lf
}

func progressSupported() bool {
	return slices.ContainsFunc(progressTerminals, func(program string) bool {
		return strings.EqualFold(program, Program)
	})
}

func StartProgress() string {
	if !progressSupported() {
		return ""
	}

	return startProgress
}

func SetProgress(percentage int) string {
	if !progressSupported() {
		return ""
	}

	return fmt.Sprintf(setProgress, percentage)
}

func StopProgress() string {
	if !progressSupported() {
		return ""
	}

	return endProgress
}

func Write(background, foreground color.Ansi, txt string) {
	if txt == "" {
		return
	}

	backgroundColor, foregroundColor = asAnsiColors(background, foreground)

	// default to white foreground
	if foregroundColor.IsEmpty() {
		foregroundColor = Colors.ToAnsi("white", false)
	}

	// reset gradient state left over from a previous Write call
	bgGradientCells, fgGradientCells = nil, nil
	cellIndex = 0
	gradientRenderCells = 0

	// isTransparent is per-segment state: a previous Write's transparent rendering
	// must not suppress gradient stamping (or trigger a spurious transparentEnd in
	// endColorOverride) for this one.
	isTransparent = false

	// asAnsiColors resolves an inverted background (transparent foreground) with a
	// foreground code for writeTransparentStart; a gradient bypasses that conversion,
	// so collapse it here and take the regular transparent path: a valid gradient
	// shows its first stop (this glyph is the segment's left edge), an invalid one
	// its last stop, matching the solid color the body falls back to.
	if foregroundColor.IsTransparent() && backgroundColor.IsGradient() {
		if color.GradientCells(backgroundColor, 1, Colors, false, CurrentColors, ParentColors) != nil {
			backgroundColor = collapseGradientFirst(backgroundColor, false)
		} else {
			backgroundColor = collapseGradientLast(backgroundColor, false)
		}
	}

	bgGradient := backgroundColor.IsGradient()
	fgGradient := foregroundColor.IsGradient()

	// validate if we start with a color override
	match := scanAnchor(txt)
	body := txt[len(match.Anchor):]

	if match.ok && match.Anchor != hyperLinkStart {
		colorOverride := true
		for _, style := range knownStyles {
			if match.Anchor != style.AnchorStart {
				continue
			}

			writeEscapedAnsiString(style.Start)
			colorOverride = false
		}

		if colorOverride {
			currentColor.Add(asAnsiColors(color.Ansi(match.BG), color.Ansi(match.FG)))
		}
	}

	// a gradient needs the segment's visible cell count before anything streams,
	// so GradientCells can hand back one color per cell up front.
	if bgGradient || fgGradient {
		cells := countVisibleCells(body, match.Anchor == hyperLinkStart)
		gradientRenderCells = cells

		if bgGradient {
			bgGradientCells = color.GradientCells(backgroundColor, cells, Colors, true, CurrentColors, ParentColors)
			if bgGradientCells == nil {
				// invalid gradient (e.g. a single resolvable stop): collapse to the
				// LAST stop so the body matches the engine's width collapse and the
				// last-stop edges separators and parent keywords already render.
				backgroundColor = collapseGradientLast(backgroundColor, true)
				bgGradient = false
			}
		}

		if fgGradient {
			fgGradientCells = color.GradientCells(foregroundColor, cells, Colors, false, CurrentColors, ParentColors)
			if fgGradientCells == nil {
				foregroundColor = collapseGradientLast(foregroundColor, false)
				fgGradient = false
			}
		}
	}

	writeSegmentColors()

	// print the hyperlink part AFTER the coloring
	if match.ok && match.Anchor == hyperLinkStart {
		isHyperlink = true
		builder.WriteString(formats.HyperlinkStart)
	}

	txt = body
	textLen = len(txt)

	if bgGradient || fgGradient {
		writeBodyGradient(txt, background)
	} else {
		writeBody(txt, background)
	}

	// reset colors
	writeEscapedAnsiString(resetStyle.End)

	// pop last color from the stack
	currentColor.Pop()
}

// writeBody streams txt's visible runes, style/color overrides and hyperlink
// tokens to the builder. It is the fast path used whenever neither channel of
// the segment being written is a gradient: no per-rune branching beyond what
// existed before gradients were added.
func writeBody(txt string, background color.Ansi) {
	hyperlinkTextPosition := 0

	for i := 0; i < len(txt); {
		s, size := utf8.DecodeRuneInString(txt[i:])

		// ignore everything which isn't overriding
		if s != '<' {
			write(s)
			i += size
			continue
		}

		// color/end overrides first
		match := scanAnchor(txt[i:])
		if match.ok {
			// check for hyperlinks first
			switch match.Anchor {
			case hyperLinkStart:
				isHyperlink = true
				i += len(match.Anchor)
				builder.WriteString(formats.HyperlinkStart)
				continue
			case hyperLinkText:
				isHyperlink = false
				i += len(match.Anchor)
				hyperlinkTextPosition = i
				builder.WriteString(formats.HyperlinkCenter)
				continue
			case hyperLinkTextEnd:
				// this implies there's no text in the hyperlink
				if hyperlinkTextPosition == i {
					builder.WriteString("link")
					length += 4
				}
				i += len(match.Anchor)
				continue
			case hyperLinkEnd:
				i += len(match.Anchor)
				builder.WriteString(formats.HyperlinkEnd)
				continue
			case empty:
				i += len(match.Anchor)
				continue
			}

			i = writeAnchorOverride(match, background, i)
			continue
		}

		write(s)
		i += size
	}
}

// writeBodyGradient is writeBody's counterpart for when at least one channel is
// a gradient. It stamps the interpolated color for the active, non-overridden
// channel(s) before every visible rune (and the hyperlink no-text fallback),
// advancing cellIndex in lockstep with countVisibleCells's pre-pass count.
func writeBodyGradient(txt string, background color.Ansi) {
	hyperlinkTextPosition := 0

	for i := 0; i < len(txt); {
		s, size := utf8.DecodeRuneInString(txt[i:])

		// ignore everything which isn't overriding
		if s != '<' {
			writeVisibleRune(s)
			i += size
			continue
		}

		// color/end overrides first
		match := scanAnchor(txt[i:])
		if match.ok {
			// check for hyperlinks first
			switch match.Anchor {
			case hyperLinkStart:
				isHyperlink = true
				i += len(match.Anchor)
				builder.WriteString(formats.HyperlinkStart)
				continue
			case hyperLinkText:
				isHyperlink = false
				i += len(match.Anchor)
				hyperlinkTextPosition = i
				builder.WriteString(formats.HyperlinkCenter)
				continue
			case hyperLinkTextEnd:
				// this implies there's no text in the hyperlink
				if hyperlinkTextPosition == i {
					stampGradient()
					builder.WriteString("link")
					length += 4
					cellIndex += 4
				}
				i += len(match.Anchor)
				continue
			case hyperLinkEnd:
				i += len(match.Anchor)
				builder.WriteString(formats.HyperlinkEnd)
				continue
			case empty:
				i += len(match.Anchor)
				continue
			}

			i = writeAnchorOverride(match, background, i)
			continue
		}

		writeVisibleRune(s)
		i += size
	}
}

func Len() int {
	return length
}

func String() (string, int) {
	defer func() {
		length = 0
		builder.Reset()

		isTransparent = false
		isInvisible = false

		bgGradientCells, fgGradientCells = nil, nil
		cellIndex = 0

		// the parent stack is scoped to one block; each new block starts a
		// fresh ancestor chain. Slicing to zero keeps the backing array so
		// same-size blocks (the common case) push without reallocating.
		ParentColors = ParentColors[:0]
	}()

	return builder.String(), length
}

func writeEscapedAnsiString(txt string) {
	if Plain {
		return
	}

	if len(escapePrefix) != 0 {
		builder.WriteString(escapePrefix)
	}

	builder.WriteString(txt)

	if len(escapeSuffix) != 0 {
		builder.WriteString(escapeSuffix)
	}
}

// writeEscapedAnsiParts writes prefix+payload+suffix wrapped in the shell escape
// sequence, avoiding the intermediate string concatenation that a
// fmt.Sprintf(colorise/transparentStart, ...) call would otherwise require.
func writeEscapedAnsiParts(prefix string, payload color.Ansi, suffix string) {
	if Plain {
		return
	}

	if len(escapePrefix) != 0 {
		builder.WriteString(escapePrefix)
	}

	builder.WriteString(prefix)
	builder.WriteString(payload.String())
	builder.WriteString(suffix)

	if len(escapeSuffix) != 0 {
		builder.WriteString(escapeSuffix)
	}
}

// writeColorise writes the equivalent of fmt.Sprintf(colorise, c) wrapped in
// the shell escape sequence, without allocating an intermediate string.
// An empty payload would emit a bare \x1b[m (a full SGR reset) and a raw
// gradient string would emit garbage; both degrade to writing nothing so a
// missed guard upstream costs a color, never corrupted output.
func writeColorise(c color.Ansi) {
	if c.IsEmpty() || c.IsGradient() {
		return
	}

	writeEscapedAnsiParts(colorisePrefix, c, coloriseSuffix)
}

// writeTransparentStart writes the equivalent of fmt.Sprintf(transparentStart, c)
// wrapped in the shell escape sequence, without allocating an intermediate string.
// The empty/gradient guard mirrors writeColorise: \x1b[;49m\x1b[7m would run
// reverse video against default colors instead of the intended payload.
func writeTransparentStart(c color.Ansi) {
	if c.IsEmpty() || c.IsGradient() {
		return
	}

	writeEscapedAnsiParts(transparentStartPrefix, c, transparentStartSuffix)
}

func write(s rune) {
	if isInvisible {
		return
	}

	// segment content (directory names, git metadata, environment variables,
	// command output) is potentially attacker-controlled and never passes through
	// this function when it's Oh My Posh's own styling; drop C0/C1 control runes
	// (ESC, BEL, CSI, OSC, ...) so they can't be interpreted as escape sequences
	// by the terminal, including inside a hyperlink target.
	if isControlRune(s) {
		return
	}

	if isHyperlink {
		builder.WriteRune(s)
		return
	}

	// UNSOLVABLE: When "Interactive" is true, the prompt length calculation in Bash/Zsh can be wrong, since the final string expansion is done by shells.
	length += runewidth.RuneWidth(s)
	// length += utf8.RuneCountInString(string(s))

	if !Interactive && !Plain {
		escaped, shouldEscape := formats.EscapeSequences[s]
		if shouldEscape {
			builder.WriteString(escaped)
			return
		}
	}

	builder.WriteRune(s)
}

// isControlRune reports whether s is a C0 (0x00-0x1F), DEL (0x7F), or C1
// (0x80-0x9F) control character. These are the bytes a terminal can interpret
// as the start of an escape sequence (ESC, BEL, CSI, OSC, ...); no legitimate
// rendered segment content needs them. '\n' is exempt: Oh My Posh itself
// prepends a literal newline ahead of the transient prompt (see
// Engine.getNewline), and unlike ESC/BEL/CSI/OSC a bare LF can't be
// interpreted as the start of an escape sequence.
func isControlRune(s rune) bool {
	if s == '\n' {
		return false
	}

	return s <= 0x1f || (s >= 0x7f && s <= 0x9f)
}

// writeVisibleRune stamps the active gradient color(s) for the current cell
// before writing s, then advances cellIndex by s's rune width. It is only
// called from writeBodyGradient, so isInvisible/isHyperlink runes are
// excluded from stamping and the index exactly like write() excludes them
// from length.
func writeVisibleRune(s rune) {
	visible := !isInvisible && !isHyperlink

	if visible {
		stampGradient()
	}

	write(s)

	if visible {
		cellIndex += runewidth.RuneWidth(s)
	}
}

// stampGradient writes the truecolor/256-color escape for the current cell of
// each channel that has a gradient AND is not currently suppressed by an
// inline override. A channel counts as overridden when the color history's
// top entry (or the segment base, when the history is empty) no longer
// matches the channel's original gradient value; endColorOverride restores
// that match on `</>`, which is what makes stamping resume automatically.
func stampGradient() {
	// transparent (reverse video) rendering collapses a gradient to a single edge
	// color; stamping a background escape here would corrupt the inverted state.
	if isTransparent {
		return
	}

	if len(bgGradientCells) != 0 && activeBackground() == backgroundColor {
		writeColorise(bgGradientCells[clampCellIndex(len(bgGradientCells))])
	}

	if len(fgGradientCells) != 0 && activeForeground() == foregroundColor {
		writeColorise(fgGradientCells[clampCellIndex(len(fgGradientCells))])
	}
}

// clampCellIndex guards against cellIndex reaching n on a trailing zero-width
// rune (e.g. a newline after the last printable cell), which would otherwise
// index one past the end of a gradient's cell slice.
func clampCellIndex(n int) int {
	if cellIndex >= n {
		return n - 1
	}

	return cellIndex
}

// activeBackground/activeForeground return the color currently in effect for
// each channel: the top of the override history, or the segment base color
// when no override is active.
func activeBackground() color.Ansi {
	if bg := currentColor.Background(); !bg.IsEmpty() {
		return bg
	}

	return backgroundColor
}

func activeForeground() color.Ansi {
	if fg := currentColor.Foreground(); !fg.IsEmpty() {
		return fg
	}

	return foregroundColor
}

// gradientCell resolves c to the stamped gradient color at the current cell when c
// is the segment's own gradient for either channel (a `background`/`foreground`
// keyword override resolves to exactly that string), converting the code to the
// requested channel. This is what makes a trailing `<background,transparent>` cap
// follow the gradient to its last stop instead of collapsing to the first.
func gradientCell(c color.Ansi, isBackground bool) (color.Ansi, bool) {
	var cell color.Ansi

	switch {
	case c == backgroundColor && len(bgGradientCells) != 0:
		cell = bgGradientCells[clampCellIndex(len(bgGradientCells))]
	case c == foregroundColor && len(fgGradientCells) != 0:
		cell = fgGradientCells[clampCellIndex(len(fgGradientCells))]
	default:
		return "", false
	}

	return cell.ToChannel(isBackground), true
}

// collapseGradientEdge resolves a gradient override to the stamped color at the
// current cell when it matches the segment's own gradient, and to its first stop
// otherwise (foreign gradients, invalid context).
func collapseGradientEdge(c color.Ansi, isBackground bool) color.Ansi {
	if cell, ok := gradientCell(c, isBackground); ok {
		return cell
	}

	return collapseGradientFirst(c, isBackground)
}

// collapseGradientFirst resolves a gradient's first stop through the same
// Colors.Resolve/ToAnsi pipeline asAnsiColors applies to a literal color,
// producing a ready-to-print ANSI code. Used wherever a gradient must
// collapse to a single edge color instead of per-cell rendering: an invalid
// gradient (color.GradientCells returned nil) and the transparent-foreground
// paths, which never render gradients per cell.
func collapseGradientFirst(c color.Ansi, isBackground bool) color.Ansi {
	return collapseGradientStop(c.GradientFirst(), isBackground)
}

// collapseGradientLast is collapseGradientFirst's right-edge counterpart, used for
// the invalid-gradient fallback so the body matches the last-stop color the engine's
// width collapse and every edge consumer (separators, parent keywords) already use.
// Uses gradientRenderCells so a dark-gradient/light-gradient edge matches the actual
// last cell GradientCells rendered THIS segment's body as (see GradientLastForCells).
func collapseGradientLast(c color.Ansi, isBackground bool) color.Ansi {
	return collapseGradientStop(c.GradientLastForCells(gradientRenderCells), isBackground)
}

func collapseGradientStop(stop color.Ansi, isBackground bool) color.Ansi {
	// a syntactically invalid gradient has no stop to fall back to; return
	// an empty color rather than letting the raw string reach an escape sequence.
	if stop.IsGradient() {
		return ""
	}

	// a keyword stop (parentBackground, ...) resolves against the segment context,
	// like GradientCells does for per-cell rendering; without this the keyword string
	// reaches ToAnsi, fails, and the glyph renders colorless.
	stop = stop.Resolve(CurrentColors, ParentColors)
	if stop.IsTransparent() {
		return ""
	}

	if resolved, err := Colors.Resolve(stop); err == nil {
		stop = resolved
	}

	resolved := Colors.ToAnsi(stop, isBackground)

	// a stop that RESOLVED to a gradient (palette entry or keyword whose target is
	// itself a gradient) must never leave as a raw string; degrade to no color.
	if resolved.IsGradient() {
		return ""
	}

	return resolved
}

// countVisibleCells is the pre-pass a gradient channel needs before streaming
// starts: it walks txt with the exact same tokenization rules as
// writeBody/writeBodyGradient (scanAnchor, the hyperlink tokens, the "link"
// no-text fallback) and sums runewidth.RuneWidth over every rune write()
// would count toward length, so color.GradientCells gets the right cell
// count and the streaming loop's cellIndex never drifts from it.
// startHyperlink mirrors the loop having already consumed a leading
// hyperLinkStart anchor before txt begins.
func countVisibleCells(txt string, startHyperlink bool) int {
	cells := 0
	hyperlink := startHyperlink
	invisible := false
	hyperlinkTextPosition := 0

	// isStyleOrReset reports whether the anchor is a style tag or reset, which
	// never change the invisible state in writeAnchorOverride.
	isStyleOrReset := func(anchor string) bool {
		if anchor == resetStyle.AnchorEnd {
			return true
		}

		for _, style := range knownStyles {
			if anchor == style.AnchorStart || anchor == style.AnchorEnd {
				return true
			}
		}

		return false
	}

	for i := 0; i < len(txt); {
		s, size := utf8.DecodeRuneInString(txt[i:])

		if s != '<' {
			if !hyperlink && !invisible {
				cells += runewidth.RuneWidth(s)
			}
			i += size
			continue
		}

		match := scanAnchor(txt[i:])
		if !match.ok {
			if !hyperlink && !invisible {
				cells += runewidth.RuneWidth(s)
			}
			i += size
			continue
		}

		switch match.Anchor {
		case hyperLinkStart:
			hyperlink = true
		case hyperLinkText:
			hyperlink = false
			hyperlinkTextPosition = i + len(match.Anchor)
		case hyperLinkTextEnd:
			if hyperlinkTextPosition == i {
				cells += 4
			}
		case empty:
			// no state change
		default:
			// a color override anchor sets the invisible state exactly like
			// writeAnchorOverride's isInvisible: both channels transparent hide the
			// runes from write() and from cellIndex, so they must not be counted.
			// This models the literal `<transparent,transparent>` form; a keyword
			// that RESOLVES to transparent is not visible to this pre-pass.
			if !isStyleOrReset(match.Anchor) {
				invisible = match.FG == string(color.Transparent) && match.BG == string(color.Transparent)
			}
		}

		i += len(match.Anchor)
	}

	return cells
}

// VisibleCells returns the number of visible cells txt would render, using the exact same
// tokenization rules as Write's own pre-pass (scanAnchor, hyperlink tokens, the "link" no-text
// fallback): it strips a leading hyperlink anchor the same way Write does before delegating to
// countVisibleCells, so a caller that needs a segment's width before Write runs (e.g. the prompt
// engine's gradient minimum-width collapse) gets the identical count Write itself would use.
func VisibleCells(txt string) int {
	match := scanAnchor(txt)
	body := txt[len(match.Anchor):]

	return countVisibleCells(body, match.Anchor == hyperLinkStart)
}

func writeSegmentColors() {
	// use correct starting colors
	bg := backgroundColor
	fg := foregroundColor
	if !currentColor.Background().IsEmpty() {
		bg = currentColor.Background()
	}
	if !currentColor.Foreground().IsEmpty() {
		fg = currentColor.Foreground()
	}

	// ignore processing fully transparent colors
	isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if isInvisible {
		return
	}

	switch {
	case fg.IsTransparent() && len(BackgroundColor) != 0:
		background := Colors.ToAnsi(BackgroundColor, false)
		writeColorise(background)

		invertBg := bg
		if invertBg.IsGradient() {
			invertBg = collapseGradientEdge(invertBg, false)
		}
		writeColorise(invertBg.ToForeground())
	case fg.IsTransparent() && !bg.IsEmpty():
		isTransparent = true

		transparentBg := bg
		if transparentBg.IsGradient() {
			// the transparentStart format takes a foreground code, matching how
			// asAnsiColors resolves an inverted (transparent foreground) background.
			transparentBg = collapseGradientEdge(transparentBg, false)
		}
		writeTransparentStart(transparentBg)
	default:
		// the segment's own gradient channel is stamped per cell by
		// writeBodyGradient/stampGradient instead of once here; any other gradient
		// (e.g. a <background,...> anchor override) collapses to its first stop so
		// the raw "linear-gradient(...)" value never reaches an escape sequence.
		if !bg.IsEmpty() && !bg.IsTransparent() {
			switch {
			case !bg.IsGradient():
				writeColorise(bg)
			case len(bgGradientCells) == 0 || bg != backgroundColor:
				writeColorise(collapseGradientEdge(bg, true))
			}
		}

		if !fg.IsEmpty() && !fg.IsTransparent() {
			switch {
			case !fg.IsGradient():
				writeColorise(fg)
			case len(fgGradientCells) == 0 || fg != foregroundColor:
				writeColorise(collapseGradientEdge(fg, false))
			}
		}
	}

	// set current colors
	currentColor.Add(bg, fg)
}

func writeAnchorOverride(match anchorMatch, background color.Ansi, i int) int {
	position := i
	// check color reset first
	if match.Anchor == resetStyle.AnchorEnd {
		return endColorOverride(position)
	}

	position += len(match.Anchor)

	for _, style := range knownStyles {
		if style.AnchorEnd == match.Anchor {
			writeEscapedAnsiString(style.End)
			return position
		}
		if style.AnchorStart == match.Anchor {
			writeEscapedAnsiString(style.Start)
			return position
		}
	}

	bgColor := color.Ansi(match.BG)
	fgColor := color.Ansi(match.FG)

	if fgColor.IsTransparent() && bgColor.IsEmpty() {
		bgColor = background
	}

	bg, fg := asAnsiColors(bgColor, fgColor)

	// ignore processing fully transparent colors
	isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if isInvisible {
		return position
	}

	// make sure we have colors
	if fg.IsEmpty() {
		fg = foregroundColor
	}
	if bg.IsEmpty() {
		bg = backgroundColor
	}

	currentColor.Add(bg, fg)

	if currentColor.Foreground().IsTransparent() && len(BackgroundColor) != 0 {
		background := Colors.ToAnsi(BackgroundColor, false)
		writeColorise(background)

		invertBg := currentColor.Background()
		if invertBg.IsGradient() {
			invertBg = collapseGradientEdge(invertBg, false)
		}
		writeColorise(invertBg.ToForeground())
		return position
	}

	if currentColor.Foreground().IsTransparent() && !currentColor.Background().IsTransparent() {
		isTransparent = true

		transparentBg := currentColor.Background()
		if transparentBg.IsGradient() {
			// the transparentStart format takes a foreground code, matching how
			// asAnsiColors resolves an inverted (transparent foreground) background.
			transparentBg = collapseGradientEdge(transparentBg, false)
		}
		writeTransparentStart(transparentBg)
		return position
	}

	if currentColor.Background() != backgroundColor {
		// end the colors in case we have a transparent background
		switch {
		case currentColor.Background().IsTransparent():
			writeEscapedAnsiString(backgroundEnd)
		case currentColor.Background().IsGradient():
			// an override resolving to a gradient (e.g. a <background,...> anchor in a
			// gradient segment) collapses to its first stop; a matching gradient is
			// handled by stamping and never reaches this branch.
			writeColorise(collapseGradientEdge(currentColor.Background(), true))
		default:
			writeColorise(currentColor.Background())
		}
	}

	if currentColor.Foreground() != foregroundColor {
		fg := currentColor.Foreground()
		if fg.IsGradient() {
			fg = collapseGradientEdge(fg, false)
		}

		writeColorise(fg)
	}

	return position
}

func endColorOverride(position int) int {
	// make sure to reset the colors if needed
	position += len(resetStyle.AnchorEnd)

	// do not restore colors at the end of the string, we print it anyways
	if position == textLen {
		currentColor.Pop()
		return position
	}

	// reset colors to previous when we have more than 1 in stack
	// as soon as we have  more than 1, we can pop the last one
	// and print the previous override as it wasn't ended yet
	if currentColor.Len() > 1 {
		fg := currentColor.Foreground()
		bg := currentColor.Background()

		currentColor.Pop()

		previousBg := currentColor.Background()
		previousFg := currentColor.Foreground()

		if isTransparent {
			writeEscapedAnsiString(transparentEnd)
			// the transparent override has ended; without this reset stampGradient
			// stays suppressed and a gradient background never resumes stamping.
			isTransparent = false
		}

		// a gradient previousBg/previousFg is restored by stamping resuming on the
		// next visible rune, never by printing its raw "linear-gradient(...)" value.
		if previousBg != bg && !previousBg.IsGradient() {
			if previousBg.IsClear() {
				writeEscapedAnsiString(backgroundStyle.End)
			} else {
				writeColorise(previousBg)
			}
		}

		if previousFg != fg && !previousFg.IsGradient() {
			writeColorise(previousFg)
		}

		return position
	}

	// pop the last colors from the stack
	defer currentColor.Pop()

	// do not reset when colors are identical
	if currentColor.Background() == backgroundColor && currentColor.Foreground() == foregroundColor {
		return position
	}

	if isTransparent {
		writeEscapedAnsiString(transparentEnd)
	}

	if backgroundColor.IsClear() {
		writeEscapedAnsiString(backgroundStyle.End)
	}

	// a gradient backgroundColor/foregroundColor is restored by stamping resuming
	// on the next visible rune, never printed here directly.
	if currentColor.Background() != backgroundColor && !backgroundColor.IsClear() && !backgroundColor.IsGradient() {
		writeColorise(backgroundColor)
	}

	if (currentColor.Foreground() != foregroundColor || isTransparent) && !foregroundColor.IsClear() && !foregroundColor.IsGradient() {
		writeColorise(foregroundColor)
	}

	isTransparent = false
	return position
}

func asAnsiColors(background, foreground color.Ansi) (color.Ansi, color.Ansi) {
	if background == "" {
		background = color.Background
	}

	if foreground == "" {
		foreground = color.Foreground
	}

	background = background.Resolve(CurrentColors, ParentColors)
	foreground = foreground.Resolve(CurrentColors, ParentColors)

	if bg, err := Colors.Resolve(background); err == nil {
		background = bg
	}

	if fg, err := Colors.Resolve(foreground); err == nil {
		foreground = fg
	}

	inverted := foreground == color.Transparent && len(background) != 0

	background = Colors.ToAnsi(background, !inverted)
	foreground = Colors.ToAnsi(foreground, false)

	return background, foreground
}

func trimAnsi(txt string) string {
	if txt == "" || !strings.Contains(txt, "\x1b") {
		return txt
	}
	return regex.ReplaceAllString(AnsiRegex, txt, "")
}
