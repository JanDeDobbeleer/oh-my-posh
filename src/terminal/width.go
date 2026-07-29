package terminal

import (
	"unicode"

	"golang.org/x/text/width"
)

// runeCells reports how many terminal cells a rune paints: 0, 1 or 2. Every length the writer
// computes is a sum of these - prompt length (which decides where a right-aligned block, an
// rprompt or a filler lands), and the per-cell index a gradient interpolates across.
//
// This replaces github.com/mattn/go-runewidth, which did the same job from its own Unicode
// tables. It was dropped for one reason: v0.0.27 eagerly builds a 2.2 MB lookup table in package
// init - 2.23 million calls before main runs, measured at 26 ms. oh-my-posh is a short-lived
// process that renders one prompt and exits, so that init was ~10% of a 240 ms render, on every
// prompt, twice per keystroke on shells that spawn a transient prompt too. The table it builds is
// a cache: measured over a realistic 80-rune prompt it saved 48 ns per render against the
// library's own uncached path, so break-even was somewhere past half a million renders in one
// process. It also could not be turned off - Condition.RuneWidth only consults the table when
// StrictEmojiNeutral is set, but init builds it either way.
//
// The tables below come from the standard library's own unicode package and golang.org/x/text,
// both already linked, so this costs no init at all.
//
// Verified against go-runewidth v0.0.27 over all 1,114,112 code points: 2,648 disagree, of which
// 2,048 are UTF-16 surrogates (unreachable - utf8.DecodeRuneInString yields RuneError before this
// is ever called), ~330 are wide symbol blocks nobody puts in a prompt (I Ching hexagrams, Tai
// Xuan Jing, counting rods, Tangut), and ~150 are format/tag characters where returning 0 is the
// better answer. The genuine gap is ~70 recently-assigned combining marks that x/text's Unicode
// vintage does not yet classify as marks; they render one cell wide instead of zero. Every rune a
// prompt actually contains agrees exactly: emoji including flag sequences, CJK, Nerd Font private
// use glyphs, powerline separators, box drawing, accented Latin, common combining marks, Cyrillic
// and Greek.
func runeCells(r rune) int {
	if r < 0 || r > unicode.MaxRune {
		return 0
	}

	// Fast path. Everything below U+0300 is one cell except the C0/C1 control ranges, and it is
	// almost everything a prompt is made of, so this answers without touching a table.
	if r < 0x0300 {
		if r < 0x20 || (r >= 0x7F && r < 0xA0) {
			return 0
		}

		return 1
	}

	// Combining marks (Mn/Me) attach to the preceding cell, and format characters (Cf) - zero
	// width joiners, bidi controls, the tag characters inside emoji flag sequences - paint
	// nothing at all.
	if unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf) {
		return 0
	}

	// East Asian Wide and Fullwidth are the two properties a terminal renders double width. The
	// remaining kinds (Narrow, Halfwidth, Ambiguous, Neutral) are all single width here, which is
	// what go-runewidth's own EastAsianWidth: false setting meant.
	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	default:
		return 1
	}
}
