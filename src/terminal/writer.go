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
)

type style struct {
	AnchorStart string
	AnchorEnd   string
	Start       string
	End         string
}

// colorState bundles the segment-scoped color/gradient bookkeeping the ANSI
// emission functions need, passed explicitly instead of read as bare package
// globals: an encoder driven from something other than Write's own prologue
// (a run stream, eventually) needs these as inputs, not ambient state.
//
// currentColor is the override history color/style anchors push onto.
//
// bgGradientCells/fgGradientCells hold one ready-to-print ANSI code per
// visible cell of the segment being written, populated by color.GradientCells
// when the corresponding channel is a gradient. cellIndex is the shared
// cursor into both slices, advanced once per visible rune regardless of which
// channel(s) stamp. See stampGradient/writeVisibleRune.
//
// backgroundColor/foregroundColor are the segment's resolved SGR pair;
// backgroundColorSource/foregroundColorSource are their pre-ToAnsi source
// form (see asAnsiColorsWithSource): a #RRGGBB hex, a colour name, accent,
// transparent, or a gradient definition. Set alongside backgroundColor/
// foregroundColor in Write, read by writeSegmentColors as the fallback
// source when no currentColor override is active.
//
// isTransparent/isInvisible are the two suppression flags
// writeSegmentColors/writeAnchorOverride flip; endColorOverride reads and
// clears isTransparent, stampGradient reads isTransparent, and write reads
// isInvisible.
type colorState struct {
	backgroundColor       color.Ansi
	foregroundColor       color.Ansi
	backgroundColorSource color.Ansi
	foregroundColorSource color.Ansi
	currentColor          color.History
	bgGradientCells       []color.Ansi
	fgGradientCells       []color.Ansi
	bgGradientRGB         []color.RGB
	fgGradientRGB         []color.RGB
	cellIndex             int
	// gradientRenderCells is the segment currently being written's visible cell count,
	// set once cells is known (see Write). collapseGradientLast reads it so a
	// dark-gradient/light-gradient color override edge mid-body matches the same shade
	// GradientCells rendered the segment's actual last cell as (see GradientLastForCells).
	// Zero (its reset value) falls back to GradientLast's gentlest single-step shade.
	gradientRenderCells int
	isTransparent       bool
	isInvisible         bool
}

var (
	// knownStyles is an ARRAY, not a slice: it is only ever ranged over and indexed
	// (never appended to), and a fixed-size array lets runAttributeSlots be derived
	// from it as a compile-time constant (len of an array literal is a constant
	// expression; len of a slice var is not).
	knownStyles = [...]*style{
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

	// colorsState is the single instance of the segment-scoped color/gradient
	// bookkeeping the ANSI emission functions (writeSegmentColors,
	// writeAnchorOverride, endColorOverride, stampGradient, write,
	// writeVisibleRune, and their helpers) consume. Write populates it per
	// call; String's defer resets the fields a fresh Write call doesn't
	// unconditionally overwrite. See colorState for what each field carries.
	colorsState colorState
	textLen     int

	isHyperlink bool

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

	// DECSCUSR (CSI n SP q) cursor style names, matching xterm's ctlseqs terminology.
	BlinkingBlock     = "blinking_block"
	SteadyBlock       = "steady_block"
	BlinkingUnderline = "blinking_underline"
	SteadyUnderline   = "steady_underline"
	BlinkingBar       = "blinking_bar"
	SteadyBar         = "steady_bar"

	// DefaultSteady and DefaultBlinking reset DECSCUSR to the terminal's own
	// default shape (CSI 0 SP q) and only toggle blink (CSI ? 12 h/l), so
	// terminal-specific shapes DECSCUSR can't select - Windows Terminal's
	// vintage, double-underscore and empty-box cursors among them - stay
	// under the user's terminal profile setting instead of being overridden.
	DefaultSteady   = "default_steady"
	DefaultBlinking = "default_blinking"

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

// anchorKind classifies a scanned anchor token into the cases writeBody,
// writeBodyGradient and countVisibleCells all branch on once a `<` is found.
// anchorNone means no anchor token starts at txt[i] (a literal '<'); the three
// hyperlink-transition kinds and anchorEmpty mirror the named tokens exactly;
// anchorOverride is everything else, i.e. a color/style override anchor
// (including `</>`), destined for writeAnchorOverride.
type anchorKind int

const (
	anchorNone anchorKind = iota
	anchorHyperlinkStart
	anchorHyperlinkText
	anchorHyperlinkTextEnd
	anchorHyperlinkEnd
	anchorEmpty
	anchorOverride
)

// classifyAnchor scans txt[i:] (txt[0] == '<', checked by the caller's fast
// path) and classifies the result. It is the single piece of logic shared by
// writeBody, writeBodyGradient and countVisibleCells: what each does with a
// given kind, and how it advances its own cursor, stays with the caller — see
// the "Index ownership" note on writeAnchorOverride's callers.
func classifyAnchor(txt string) (anchorMatch, anchorKind) {
	match := scanAnchor(txt)
	if !match.ok {
		return match, anchorNone
	}

	switch match.Anchor {
	case hyperLinkStart:
		return match, anchorHyperlinkStart
	case hyperLinkText:
		return match, anchorHyperlinkText
	case hyperLinkTextEnd:
		return match, anchorHyperlinkTextEnd
	case hyperLinkEnd:
		return match, anchorHyperlinkEnd
	case empty:
		return match, anchorEmpty
	default:
		return match, anchorOverride
	}
}

// leadingAnchorInvisible reports whether match is a leading anchor that Write's
// prologue (and VisibleCells' mirror of it) would consume as a fully transparent
// color override, i.e. a literal `<transparent,transparent>` anchor. It excludes
// hyperlink starts and style anchors exactly like Write's prologue already does
// when deciding whether to treat the anchor as a color override at all, and uses
// the same literal FG/BG comparison countVisibleCells' own body scan uses (a
// keyword that RESOLVES to transparent is not visible to this pre-pass).
func leadingAnchorInvisible(match anchorMatch) bool {
	if !match.ok || match.Anchor == hyperLinkStart {
		return false
	}

	for _, style := range knownStyles[:] {
		if match.Anchor == style.AnchorStart {
			return false
		}
	}

	return match.FG == string(color.Transparent) && match.BG == string(color.Transparent)
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

	userName = stripControlRunes(userName)
	hostName = stripControlRunes(hostName)
	pwd = stripControlRunes(pwd)

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

// decscusrCodes maps the CursorStyle option values to their DECSCUSR (CSI n SP q) parameter.
var decscusrCodes = map[string]int{
	BlinkingBlock:     1,
	SteadyBlock:       2,
	BlinkingUnderline: 3,
	SteadyUnderline:   4,
	BlinkingBar:       5,
	SteadyBar:         6,
}

// SetCursorStyle renders the DECSCUSR sequence for a known cursor style, wrapped in the
// shell's escape sequence so it doesn't confuse readline-based shells about the cursor
// column. Unlike segment content this is Oh My Posh's own trusted output, so it's exempt
// from the control-rune stripping write() applies to potentially attacker-controlled text.
func SetCursorStyle(style string) string {
	if Plain {
		return ""
	}

	switch style {
	case DefaultSteady:
		return fmt.Sprintf(formats.Escape, "\x1b[0 q\x1b[?12l")
	case DefaultBlinking:
		return fmt.Sprintf(formats.Escape, "\x1b[0 q\x1b[?12h")
	}

	code, ok := decscusrCodes[style]
	if !ok {
		return ""
	}

	return fmt.Sprintf(formats.Escape, fmt.Sprintf("\x1b[%d q", code))
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

	cs := &colorsState

	cs.backgroundColor, cs.foregroundColor, cs.backgroundColorSource, cs.foregroundColorSource = asAnsiColorsWithSource(background, foreground)

	// default to white foreground
	if cs.foregroundColor.IsEmpty() {
		cs.foregroundColor = Colors.ToAnsi("white", false)
		cs.foregroundColorSource = "white"
	}

	// reset gradient state left over from a previous Write call
	cs.bgGradientCells, cs.fgGradientCells = nil, nil
	cs.bgGradientRGB, cs.fgGradientRGB = nil, nil
	cs.cellIndex = 0
	cs.gradientRenderCells = 0

	// isTransparent is per-segment state: a previous Write's transparent rendering
	// must not suppress gradient stamping (or trigger a spurious transparentEnd in
	// endColorOverride) for this one.
	cs.isTransparent = false

	// isHyperlink is per-segment state too, and unlike isInvisible it is not
	// unconditionally recomputed elsewhere in Write: it is only flipped true/false
	// when a <LINK>/<TEXT> anchor is actually encountered in the body. A previous
	// Write's unbalanced <LINK> (no closing </TEXT>) would otherwise leave it true
	// forever, routing every later Write's runes through write's isHyperlink branch
	// with no length counting and no shell escaping.
	isHyperlink = false

	// asAnsiColorsWithSource resolves an inverted background (transparent foreground)
	// with a foreground code for writeTransparentStart; a gradient bypasses that conversion,
	// so collapse it here and take the regular transparent path: a valid gradient
	// shows its first stop (this glyph is the segment's left edge), an invalid one
	// its last stop, matching the solid color the body falls back to.
	if cs.foregroundColor.IsTransparent() && cs.backgroundColor.IsGradient() {
		if color.GradientCells(cs.backgroundColor, 1, Colors, false, CurrentColors, ParentColors) != nil {
			cs.backgroundColor = collapseGradientFirst(cs.backgroundColor, false)
		} else {
			cs.backgroundColor = collapseGradientLast(cs, cs.backgroundColor, false)
		}
	}

	bgGradient := cs.backgroundColor.IsGradient()
	fgGradient := cs.foregroundColor.IsGradient()

	// validate if we start with a color override
	match := scanAnchor(txt)
	body := txt[len(match.Anchor):]

	if match.ok && match.Anchor != hyperLinkStart {
		colorOverride := true
		for idx, style := range knownStyles[:] {
			if match.Anchor != style.AnchorStart {
				continue
			}

			writeEscapedAnsiString(style.Start)
			colorOverride = false

			if CaptureRuns {
				runsState.depth[idx]++
			}
		}

		if colorOverride {
			bg, fg, bgSource, fgSource := asAnsiColorsWithSource(color.Ansi(match.BG), color.Ansi(match.FG))
			cs.currentColor.Add(bg, fg, bgSource, fgSource)
		}
	}

	// a gradient needs the segment's visible cell count before anything streams,
	// so GradientCells can hand back one color per cell up front.
	if bgGradient || fgGradient {
		// leadingInvisible mirrors startHyperlink below: countVisibleCells's pre-pass
		// needs to know the consumed leading anchor left runes invisible (a literal
		// `<transparent,transparent>` override) exactly like it already needs to know
		// the anchor started a hyperlink, or it undercounts nothing hidden and
		// overcounts everything that follows. See leadingAnchorInvisible. Computed here,
		// not unconditionally in Write's prologue: it is only ever consumed by
		// countVisibleCells below, and for the common non-gradient segment starting with
		// an anchor, calling it unconditionally cost an 8-iteration knownStyles loop and
		// two string compares on every Write for nothing. VisibleCells (a separate call
		// site) is unaffected — it always needs the value, since it never knows in
		// advance whether the caller's segment is a gradient.
		leadingInvisible := leadingAnchorInvisible(match)
		cells := countVisibleCells(body, match.Anchor == hyperLinkStart, leadingInvisible)
		cs.gradientRenderCells = cells

		if bgGradient {
			cs.bgGradientCells = color.GradientCells(cs.backgroundColor, cells, Colors, true, CurrentColors, ParentColors)
			if cs.bgGradientCells == nil {
				// invalid gradient (e.g. a single resolvable stop): collapse to the
				// LAST stop so the body matches the engine's width collapse and the
				// last-stop edges separators and parent keywords already render.
				cs.backgroundColor = collapseGradientLast(cs, cs.backgroundColor, true)
				bgGradient = false
			}

			if bgGradient && CaptureRuns {
				cs.bgGradientRGB = color.GradientCellsRGB(cs.backgroundColor, cells, Colors, CurrentColors, ParentColors)
			}
		}

		if fgGradient {
			cs.fgGradientCells = color.GradientCells(cs.foregroundColor, cells, Colors, false, CurrentColors, ParentColors)
			if cs.fgGradientCells == nil {
				cs.foregroundColor = collapseGradientLast(cs, cs.foregroundColor, false)
				fgGradient = false
			}

			if fgGradient && CaptureRuns {
				cs.fgGradientRGB = color.GradientCellsRGB(cs.foregroundColor, cells, Colors, CurrentColors, ParentColors)
			}
		}
	}

	writeSegmentColors(cs, BackgroundColor)

	// print the hyperlink part AFTER the coloring
	if match.ok && match.Anchor == hyperLinkStart {
		isHyperlink = true
		writeHyperlinkEscape(formats.HyperlinkStart)
	}

	txt = body
	textLen = len(txt)

	if bgGradient || fgGradient {
		writeBodyGradient(txt, background, cs)
	} else {
		writeBody(txt, background, cs)
	}

	if CaptureRuns {
		// cut the segment's trailing run before the SGR reset below clears every
		// attribute a real terminal would carry past it; depth must reset alongside it,
		// or a style anchor left open by unbalanced markup would leak into the next
		// Write call's runs.
		flushRun()
		runsState.depth = [runAttributeSlots]uint8{}
	}

	// reset colors
	writeEscapedAnsiString(resetStyle.End)

	// pop last color from the stack
	cs.currentColor.Pop()
}

// writeBody streams txt's visible runes, style/color overrides and hyperlink
// tokens to the builder. It is the fast path used whenever neither channel of
// the segment being written is a gradient: no per-rune branching beyond what
// existed before gradients were added.
func writeBody(txt string, background color.Ansi, cs *colorState) {
	hyperlinkTextPosition := 0

	for i := 0; i < len(txt); {
		s, size := utf8.DecodeRuneInString(txt[i:])

		// ignore everything which isn't overriding
		if s != '<' {
			write(s, cs.isInvisible)
			i += size
			continue
		}

		// color/end overrides first
		match, kind := classifyAnchor(txt[i:])
		switch kind {
		case anchorNone:
			// not an anchor after all (a literal '<'); fall through to the
			// plain-rune path below.
		case anchorHyperlinkStart:
			isHyperlink = true
			i += len(match.Anchor)
			writeHyperlinkEscape(formats.HyperlinkStart)
			continue
		case anchorHyperlinkText:
			isHyperlink = false
			i += len(match.Anchor)
			hyperlinkTextPosition = i
			writeHyperlinkEscape(formats.HyperlinkCenter)
			continue
		case anchorHyperlinkTextEnd:
			// this implies there's no text in the hyperlink
			if hyperlinkTextPosition == i {
				builder.WriteString("link")
				length += 4

				if CaptureRuns {
					runsState.text.WriteString("link")
				}
			}
			i += len(match.Anchor)
			continue
		case anchorHyperlinkEnd:
			i += len(match.Anchor)
			writeHyperlinkEscape(formats.HyperlinkEnd)
			continue
		case anchorEmpty:
			i += len(match.Anchor)
			continue
		case anchorOverride:
			i = writeAnchorOverride(cs, match, background, BackgroundColor, i)
			continue
		}

		write(s, cs.isInvisible)
		i += size
	}
}

// writeBodyGradient is writeBody's counterpart for when at least one channel is
// a gradient. It stamps the interpolated color for the active, non-overridden
// channel(s) before every visible rune (and the hyperlink no-text fallback),
// advancing cellIndex in lockstep with countVisibleCells's pre-pass count.
func writeBodyGradient(txt string, background color.Ansi, cs *colorState) {
	hyperlinkTextPosition := 0

	for i := 0; i < len(txt); {
		s, size := utf8.DecodeRuneInString(txt[i:])

		// ignore everything which isn't overriding
		if s != '<' {
			writeVisibleRune(s, cs)
			i += size
			continue
		}

		// color/end overrides first
		match, kind := classifyAnchor(txt[i:])
		switch kind {
		case anchorNone:
			// not an anchor after all (a literal '<'); fall through to the
			// plain-rune path below.
		case anchorHyperlinkStart:
			isHyperlink = true
			i += len(match.Anchor)
			writeHyperlinkEscape(formats.HyperlinkStart)
			continue
		case anchorHyperlinkText:
			isHyperlink = false
			i += len(match.Anchor)
			hyperlinkTextPosition = i
			writeHyperlinkEscape(formats.HyperlinkCenter)
			continue
		case anchorHyperlinkTextEnd:
			// this implies there's no text in the hyperlink
			if hyperlinkTextPosition == i {
				stampGradient(cs)
				builder.WriteString("link")
				length += 4
				cs.cellIndex += 4

				if CaptureRuns {
					runsState.text.WriteString("link")
				}
			}
			i += len(match.Anchor)
			continue
		case anchorHyperlinkEnd:
			i += len(match.Anchor)
			writeHyperlinkEscape(formats.HyperlinkEnd)
			continue
		case anchorEmpty:
			i += len(match.Anchor)
			continue
		case anchorOverride:
			i = writeAnchorOverride(cs, match, background, BackgroundColor, i)
			continue
		}

		writeVisibleRune(s, cs)
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

		colorsState.isTransparent = false
		colorsState.isInvisible = false
		isHyperlink = false

		colorsState.bgGradientCells, colorsState.fgGradientCells = nil, nil
		colorsState.bgGradientRGB, colorsState.fgGradientRGB = nil, nil
		colorsState.cellIndex = 0
		colorsState.gradientRenderCells = 0

		// the parent stack is scoped to one block; each new block starts a
		// fresh ancestor chain. Slicing to zero keeps the backing array so
		// same-size blocks (the common case) push without reallocating.
		ParentColors = ParentColors[:0]

		if CaptureRuns {
			// mirrors the resets above: a run stream (or leftover pending style) must
			// not survive into the next block/rprompt/transient render either.
			runsState.Runs = runsState.Runs[:0]
			runsState.text.Reset()
			runsState.background, runsState.foreground = "", ""
			runsState.backgroundSource, runsState.foregroundSource = "", ""
			runsState.backgroundRGB, runsState.foregroundRGB = nil, nil
			runsState.mode = RunNormal
			runsState.attributes = [runAttributeSlots]uint8{}
			runsState.depth = [runAttributeSlots]uint8{}
			runsState.cellsAtFlush = 0
		}
	}()

	return builder.String(), length
}

// writeHyperlinkEscape writes one of the OSC 8 hyperlink wrapper sequences
// (formats.HyperlinkStart/Center/End) directly to the builder, suppressed
// entirely in Plain mode. It is not routed through writeEscapedAnsiString:
// those sequences bracket the URL and link text, which stream through write()
// separately while isHyperlink is set, so this only needs to guard on Plain,
// not shell-escape the payload.
func writeHyperlinkEscape(txt string) {
	if Plain {
		return
	}

	builder.WriteString(txt)
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

func write(s rune, isInvisible bool) {
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

	// in Plain mode, neither the OSC 8 wrappers (writeHyperlinkEscape) nor the URL
	// text between <LINK> and <TEXT> may reach the builder: the URL runes stream
	// through here while isHyperlink is set, and unlike the plain-rune path below
	// they are never counted toward length, so leaving them unguarded would let
	// invisible text corrupt every width consumer downstream.
	if isHyperlink {
		if Plain {
			return
		}

		builder.WriteRune(s)

		// Deliberately not captured. These runes are the OSC 8 target, which a terminal
		// consumes as part of the escape and never paints; length does not count them, and
		// neither does Run.Cells. Capturing them anyway put a run's Text out of step with its
		// own Cells, and an encoder that draws Text (see svg.Encode) then printed the URL
		// alongside the label - the built-in default config renders its path segment as a
		// hyperlink, so its export read "file:~/dev~/dev". The Run stream describes what is
		// painted; the URL travels in the ANSI bytes above.

		return
	}

	// UNSOLVABLE: When "Interactive" is true, the prompt length calculation in Bash/Zsh can be wrong, since the final string expansion is done by shells.
	length += runeCells(s)
	// length += utf8.RuneCountInString(string(s))

	if !Interactive && !Plain {
		escaped, shouldEscape := formats.EscapeSequences[s]
		if shouldEscape {
			builder.WriteString(escaped)

			if CaptureRuns {
				runsState.text.WriteString(escaped)
			}

			return
		}
	}

	builder.WriteRune(s)

	if CaptureRuns {
		runsState.text.WriteRune(s)
	}
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

// isOSCPayloadControlRune reports whether s is a C0 (0x00-0x1F), DEL (0x7F),
// or C1 (0x80-0x9F) control character, with no exemption for '\n': unlike
// isControlRune's rendered segment text, the values this guards are
// single-line OSC payload fields (path, username, hostname), where a raw
// newline has no legitimate use and can only corrupt the sequence.
func isOSCPayloadControlRune(s rune) bool {
	return s <= 0x1f || (s >= 0x7f && s <= 0x9f)
}

// stripControlRunes drops control runes from a whole string, for values
// that reach the terminal in one shot rather than through write's per-rune
// stream (an OSC payload field such as a path, username, or hostname).
// Kept separate from write/isControlRune, which guard the streaming render
// path instead.
func stripControlRunes(s string) string {
	if !strings.ContainsFunc(s, isOSCPayloadControlRune) {
		return s
	}

	sb := text.NewBuilder()

	for _, r := range s {
		if isOSCPayloadControlRune(r) {
			continue
		}

		sb.WriteRune(r)
	}

	return sb.String()
}

// writeVisibleRune stamps the active gradient color(s) for the current cell
// before writing s, then advances cellIndex by s's rune width. It is only
// called from writeBodyGradient, so isInvisible/isHyperlink runes are
// excluded from stamping and the index exactly like write() excludes them
// from length.
func writeVisibleRune(s rune, cs *colorState) {
	visible := !cs.isInvisible && !isHyperlink

	if visible {
		stampGradient(cs)
	}

	write(s, cs.isInvisible)

	if visible {
		cs.cellIndex += runeCells(s)
	}
}

// stampGradient writes the truecolor/256-color escape for the current cell of
// each channel that has a gradient AND is not currently suppressed by an
// inline override. A channel counts as overridden when the color history's
// top entry (or the segment base, when the history is empty) no longer
// matches the channel's original gradient value; endColorOverride restores
// that match on `</>`, which is what makes stamping resume automatically.
func stampGradient(cs *colorState) {
	// transparent (reverse video) rendering collapses a gradient to a single edge
	// color; stamping a background escape here would corrupt the inverted state.
	if cs.isTransparent {
		return
	}

	bgActive := len(cs.bgGradientCells) != 0 && activeBackground(cs) == cs.backgroundColor
	if bgActive {
		writeColorise(cs.bgGradientCells[clampCellIndex(cs, len(cs.bgGradientCells))])
	}

	fgActive := len(cs.fgGradientCells) != 0 && activeForeground(cs) == cs.foregroundColor
	if fgActive {
		writeColorise(cs.fgGradientCells[clampCellIndex(cs, len(cs.fgGradientCells))])
	}

	if !CaptureRuns || (!bgActive && !fgActive) {
		return
	}

	// every stamped cell is its own color change (stampGradient never diffs against the
	// previous cell's value), hence its own Run: flush whatever accumulated under the
	// previous cell's style before overwriting the pending background/foreground below
	// with THIS cell's, which syncPendingStyle can't derive on its own (activeBackground/
	// activeForeground return the segment's raw gradient definition, not a per-cell
	// escape or RGB).
	flushRun()
	syncPendingStyle(cs)

	if bgActive {
		idx := clampCellIndex(cs, len(cs.bgGradientCells))
		runsState.background = cs.bgGradientCells[idx]

		if idx < len(cs.bgGradientRGB) {
			runsState.backgroundRGB = &cs.bgGradientRGB[idx]
		}
	}

	if fgActive {
		idx := clampCellIndex(cs, len(cs.fgGradientCells))
		runsState.foreground = cs.fgGradientCells[idx]

		if idx < len(cs.fgGradientRGB) {
			runsState.foregroundRGB = &cs.fgGradientRGB[idx]
		}
	}
}

// clampCellIndex guards against cellIndex reaching n on a trailing zero-width
// rune (e.g. a newline after the last printable cell), which would otherwise
// index one past the end of a gradient's cell slice.
func clampCellIndex(cs *colorState, n int) int {
	if cs.cellIndex >= n {
		return n - 1
	}

	return cs.cellIndex
}

// activeBackground/activeForeground return the color currently in effect for
// each channel: the top of the override history, or the segment base color
// when no override is active.
func activeBackground(cs *colorState) color.Ansi {
	if bg := cs.currentColor.Background(); !bg.IsEmpty() {
		return bg
	}

	return cs.backgroundColor
}

func activeForeground(cs *colorState) color.Ansi {
	if fg := cs.currentColor.Foreground(); !fg.IsEmpty() {
		return fg
	}

	return cs.foregroundColor
}

// gradientCell resolves c to the stamped gradient color at the current cell when c
// is the segment's own gradient for either channel (a `background`/`foreground`
// keyword override resolves to exactly that string), converting the code to the
// requested channel. This is what makes a trailing `<background,transparent>` cap
// follow the gradient to its last stop instead of collapsing to the first.
func gradientCell(cs *colorState, c color.Ansi, isBackground bool) (color.Ansi, bool) {
	var cell color.Ansi

	switch {
	case c == cs.backgroundColor && len(cs.bgGradientCells) != 0:
		cell = cs.bgGradientCells[clampCellIndex(cs, len(cs.bgGradientCells))]
	case c == cs.foregroundColor && len(cs.fgGradientCells) != 0:
		cell = cs.fgGradientCells[clampCellIndex(cs, len(cs.fgGradientCells))]
	default:
		return "", false
	}

	return cell.ToChannel(isBackground), true
}

// collapseGradientEdge resolves a gradient override to the stamped color at the
// current cell when it matches the segment's own gradient, and to its first stop
// otherwise (foreign gradients, invalid context).
func collapseGradientEdge(cs *colorState, c color.Ansi, isBackground bool) color.Ansi {
	if cell, ok := gradientCell(cs, c, isBackground); ok {
		return cell
	}

	return collapseGradientFirst(c, isBackground)
}

// collapseGradientFirst resolves a gradient's first stop through the same
// Colors.Resolve/ToAnsi pipeline asAnsiColorsWithSource applies to a literal color,
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
// Uses cs.gradientRenderCells so a dark-gradient/light-gradient edge matches the actual
// last cell GradientCells rendered THIS segment's body as (see GradientLastForCells).
func collapseGradientLast(cs *colorState, c color.Ansi, isBackground bool) color.Ansi {
	return collapseGradientStop(c.GradientLastForCells(cs.gradientRenderCells), isBackground)
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
// no-text fallback) and sums runeCells over every rune write()
// would count toward length, so color.GradientCells gets the right cell
// count and the streaming loop's cellIndex never drifts from it.
// startHyperlink mirrors the loop having already consumed a leading
// hyperLinkStart anchor before txt begins; startInvisible mirrors it having
// instead consumed a leading fully-transparent color override anchor (see
// leadingAnchorInvisible).
func countVisibleCells(txt string, startHyperlink, startInvisible bool) int {
	cells := 0
	hyperlink := startHyperlink
	invisible := startInvisible
	hyperlinkTextPosition := 0

	// isStyleOrReset reports whether the anchor is a style tag or reset, which
	// never change the invisible state in writeAnchorOverride.
	isStyleOrReset := func(anchor string) bool {
		if anchor == resetStyle.AnchorEnd {
			return true
		}

		for _, style := range knownStyles[:] {
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
				cells += runeCells(s)
			}
			i += size
			continue
		}

		match, kind := classifyAnchor(txt[i:])
		if kind == anchorNone {
			if !hyperlink && !invisible {
				cells += runeCells(s)
			}
			i += size
			continue
		}

		switch kind {
		case anchorHyperlinkStart:
			hyperlink = true
		case anchorHyperlinkText:
			hyperlink = false
			hyperlinkTextPosition = i + len(match.Anchor)
		case anchorHyperlinkTextEnd:
			if hyperlinkTextPosition == i {
				cells += 4
			}
		case anchorHyperlinkEnd:
			// consumed like writeBody/writeBodyGradient's HyperlinkEnd case,
			// which writes formats.HyperlinkEnd without touching invisible: a
			// closing link tag is not a color override and must not flip
			// invisible back off (it used to fall to default below and be
			// treated as one, since both its FG and BG are empty).
		case anchorEmpty:
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

	return countVisibleCells(body, match.Anchor == hyperLinkStart, leadingAnchorInvisible(match))
}

func writeSegmentColors(cs *colorState, terminalBackground color.Ansi) {
	if CaptureRuns {
		flushRun()
		defer syncPendingStyle(cs)
	}

	// use correct starting colors
	bg := cs.backgroundColor
	fg := cs.foregroundColor
	bgSource := cs.backgroundColorSource
	fgSource := cs.foregroundColorSource
	if !cs.currentColor.Background().IsEmpty() {
		bg = cs.currentColor.Background()
		bgSource = cs.currentColor.BackgroundSource()
	}
	if !cs.currentColor.Foreground().IsEmpty() {
		fg = cs.currentColor.Foreground()
		fgSource = cs.currentColor.ForegroundSource()
	}

	// ignore processing fully transparent colors
	cs.isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if cs.isInvisible {
		return
	}

	switch {
	case fg.IsTransparent() && len(terminalBackground) != 0:
		background := Colors.ToAnsi(terminalBackground, false)
		writeColorise(background)

		invertBg := bg
		if invertBg.IsGradient() {
			invertBg = collapseGradientEdge(cs, invertBg, false)
		}
		writeColorise(invertBg.ToForeground())
	case fg.IsTransparent() && !bg.IsEmpty():
		cs.isTransparent = true

		transparentBg := bg
		if transparentBg.IsGradient() {
			// the transparentStart format takes a foreground code, matching how
			// asAnsiColorsWithSource resolves an inverted (transparent foreground) background.
			transparentBg = collapseGradientEdge(cs, transparentBg, false)
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
			case len(cs.bgGradientCells) == 0 || bg != cs.backgroundColor:
				writeColorise(collapseGradientEdge(cs, bg, true))
			}
		}

		if !fg.IsEmpty() && !fg.IsTransparent() {
			switch {
			case !fg.IsGradient():
				writeColorise(fg)
			case len(cs.fgGradientCells) == 0 || fg != cs.foregroundColor:
				writeColorise(collapseGradientEdge(cs, fg, false))
			}
		}
	}

	// set current colors
	cs.currentColor.Add(bg, fg, bgSource, fgSource)
}

func writeAnchorOverride(cs *colorState, match anchorMatch, background, terminalBackground color.Ansi, i int) int {
	if CaptureRuns {
		flushRun()
		defer syncPendingStyle(cs)
	}

	position := i
	// check color reset first
	if match.Anchor == resetStyle.AnchorEnd {
		return endColorOverride(cs, position)
	}

	position += len(match.Anchor)

	for idx, style := range knownStyles {
		if style.AnchorEnd == match.Anchor {
			writeEscapedAnsiString(style.End)

			if CaptureRuns && runsState.depth[idx] > 0 {
				runsState.depth[idx]--
			}

			return position
		}
		if style.AnchorStart == match.Anchor {
			writeEscapedAnsiString(style.Start)

			if CaptureRuns {
				runsState.depth[idx]++
			}

			return position
		}
	}

	bgColor := color.Ansi(match.BG)
	fgColor := color.Ansi(match.FG)

	if fgColor.IsTransparent() && bgColor.IsEmpty() {
		bgColor = background
	}

	bg, fg, bgSource, fgSource := asAnsiColorsWithSource(bgColor, fgColor)

	// ignore processing fully transparent colors
	cs.isInvisible = fg.IsTransparent() && bg.IsTransparent()
	if cs.isInvisible {
		return position
	}

	// make sure we have colors
	if fg.IsEmpty() {
		fg = cs.foregroundColor
		fgSource = cs.foregroundColorSource
	}
	if bg.IsEmpty() {
		bg = cs.backgroundColor
		bgSource = cs.backgroundColorSource
	}

	cs.currentColor.Add(bg, fg, bgSource, fgSource)

	if cs.currentColor.Foreground().IsTransparent() && len(terminalBackground) != 0 {
		background := Colors.ToAnsi(terminalBackground, false)
		writeColorise(background)

		invertBg := cs.currentColor.Background()
		if invertBg.IsGradient() {
			invertBg = collapseGradientEdge(cs, invertBg, false)
		}
		writeColorise(invertBg.ToForeground())
		return position
	}

	if cs.currentColor.Foreground().IsTransparent() && !cs.currentColor.Background().IsTransparent() {
		cs.isTransparent = true

		transparentBg := cs.currentColor.Background()
		if transparentBg.IsGradient() {
			// the transparentStart format takes a foreground code, matching how
			// asAnsiColorsWithSource resolves an inverted (transparent foreground) background.
			transparentBg = collapseGradientEdge(cs, transparentBg, false)
		}
		writeTransparentStart(transparentBg)
		return position
	}

	if cs.currentColor.Background() != cs.backgroundColor {
		// end the colors in case we have a transparent background
		switch {
		case cs.currentColor.Background().IsTransparent():
			writeEscapedAnsiString(backgroundEnd)
		case cs.currentColor.Background().IsGradient():
			// an override resolving to a gradient (e.g. a <background,...> anchor in a
			// gradient segment) collapses to its first stop; a matching gradient is
			// handled by stamping and never reaches this branch.
			writeColorise(collapseGradientEdge(cs, cs.currentColor.Background(), true))
		default:
			writeColorise(cs.currentColor.Background())
		}
	}

	if cs.currentColor.Foreground() != cs.foregroundColor {
		fg := cs.currentColor.Foreground()
		if fg.IsGradient() {
			fg = collapseGradientEdge(cs, fg, false)
		}

		writeColorise(fg)
	}

	return position
}

// endColorOverride does not flush/sync the capture bookkeeping itself: its only
// caller, writeAnchorOverride, already does that unconditionally before dispatching
// to this function (on the `</>` branch), with no style or text mutation in between.
// A future direct caller must do the same before calling this.
func endColorOverride(cs *colorState, position int) int {
	// make sure to reset the colors if needed
	position += len(resetStyle.AnchorEnd)

	// do not restore colors at the end of the string, we print it anyways
	if position == textLen {
		cs.currentColor.Pop()
		return position
	}

	// reset colors to previous when we have more than 1 in stack
	// as soon as we have  more than 1, we can pop the last one
	// and print the previous override as it wasn't ended yet
	if cs.currentColor.Len() > 1 {
		fg := cs.currentColor.Foreground()
		bg := cs.currentColor.Background()

		cs.currentColor.Pop()

		previousBg := cs.currentColor.Background()
		previousFg := cs.currentColor.Foreground()

		if cs.isTransparent {
			writeEscapedAnsiString(transparentEnd)
			// the transparent override has ended; without this reset stampGradient
			// stays suppressed and a gradient background never resumes stamping.
			cs.isTransparent = false
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
	defer cs.currentColor.Pop()

	// do not reset when colors are identical
	if cs.currentColor.Background() == cs.backgroundColor && cs.currentColor.Foreground() == cs.foregroundColor {
		return position
	}

	if cs.isTransparent {
		writeEscapedAnsiString(transparentEnd)
	}

	if cs.backgroundColor.IsClear() {
		writeEscapedAnsiString(backgroundStyle.End)
	}

	// a gradient backgroundColor/foregroundColor is restored by stamping resuming
	// on the next visible rune, never printed here directly.
	if cs.currentColor.Background() != cs.backgroundColor && !cs.backgroundColor.IsClear() && !cs.backgroundColor.IsGradient() {
		writeColorise(cs.backgroundColor)
	}

	if (cs.currentColor.Foreground() != cs.foregroundColor || cs.isTransparent) && !cs.foregroundColor.IsClear() && !cs.foregroundColor.IsGradient() {
		writeColorise(cs.foregroundColor)
	}

	cs.isTransparent = false
	return position
}

// resolveAnsiColors resolves background/foreground through keyword lookup
// (background/foreground/parentBackground/parentForeground, see color.Ansi.Resolve)
// and palette lookup (Colors.Resolve), producing each channel's SOURCE form: a
// #RRGGBB hex, a colour name, `accent`, `transparent`, or a gradient definition.
// This is the value asAnsiColors historically discarded before ToAnsi's SGR
// conversion; a later encoder needs it because SGR cannot be inverted back to
// it reliably (ToAnsi's 256-colour downgrade and base-16 codes are lossy).
func resolveAnsiColors(background, foreground color.Ansi) (color.Ansi, color.Ansi) {
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

	return background, foreground
}

// asAnsiColorsWithSource resolves background/foreground to their SGR ANSI
// codes (bg, fg — what the former asAnsiColors returned), plus each channel's
// source form from resolveAnsiColors (bgSource, fgSource). Every caller pushes
// onto currentColor, so the source rides alongside the SGR pair on the same
// color.Set entry (color.History.Add) instead of a second stack that could
// desync from it under color.TrueColor == false (see History.Add).
func asAnsiColorsWithSource(background, foreground color.Ansi) (bg, fg, bgSource, fgSource color.Ansi) {
	bgSource, fgSource = resolveAnsiColors(background, foreground)

	inverted := fgSource == color.Transparent && len(bgSource) != 0

	bg = Colors.ToAnsi(bgSource, !inverted)
	fg = Colors.ToAnsi(fgSource, false)

	return bg, fg, bgSource, fgSource
}

func trimAnsi(txt string) string {
	if txt == "" || !strings.Contains(txt, "\x1b") {
		return txt
	}
	return regex.ReplaceAllString(AnsiRegex, txt, "")
}
