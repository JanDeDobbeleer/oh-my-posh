package font

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

// buildFont assembles a minimal sfnt binary with a single name table
// holding the given entries.
func buildFont(magic string, entries []*nameEntry) []byte {
	var strData []byte
	records := make([]byte, 0, len(entries)*12)

	for _, e := range entries {
		record := make([]byte, 12)
		binary.BigEndian.PutUint16(record[0:], e.platformID)
		binary.BigEndian.PutUint16(record[2:], e.encodingID)
		binary.BigEndian.PutUint16(record[6:], uint16(e.nameID))
		binary.BigEndian.PutUint16(record[8:], uint16(len(e.value)))
		binary.BigEndian.PutUint16(record[10:], uint16(len(strData)))
		records = append(records, record...)
		strData = append(strData, e.value...)
	}

	table := make([]byte, 6)
	binary.BigEndian.PutUint16(table[2:], uint16(len(entries)))
	binary.BigEndian.PutUint16(table[4:], uint16(6+len(records)))
	table = append(table, records...)
	table = append(table, strData...)

	font := make([]byte, 12+16)
	copy(font, magic)
	binary.BigEndian.PutUint16(font[4:], 1)
	copy(font[12:], "name")
	binary.BigEndian.PutUint32(font[20:], uint32(len(font)))
	binary.BigEndian.PutUint32(font[24:], uint32(len(table)))

	return append(font, table...)
}

func utf16BE(s string) []byte {
	var b []byte
	for _, r := range s {
		b = append(b, byte(r>>8), byte(r))
	}

	return b
}

func TestNewFontMetadata(t *testing.T) {
	entries := []*nameEntry{
		// mac roman first, as name tables order entries by platform
		{platformID: platformMac, encodingID: encodingMacRoman, nameID: nameFontFamily, value: []byte("Test Family Mac")},
		{platformID: platformMicrosoft, encodingID: encodingMicrosoftUnicode, nameID: nameFontFamily, value: utf16BE("Test Family")},
		{platformID: platformMicrosoft, encodingID: encodingMicrosoftUnicode, nameID: nameFull, value: utf16BE("Test Family Regular")},
		{platformID: platformMicrosoft, encodingID: encodingMicrosoftUnicode, nameID: namePreferredFamily, value: utf16BE("Test Preferred")},
	}

	font, err := newFont("test.ttf", buildFont("\x00\x01\x00\x00", entries))
	assert.NoError(t, err)
	assert.Equal(t, "Test Family Regular", font.Name)
	assert.Equal(t, "Test Preferred", font.Family)
	// the later entry for the same name ID wins, as it did before
	assert.Equal(t, "Test Family", font.Metadata[nameFontFamily])
}

func TestNewFontFamilyFallback(t *testing.T) {
	entries := []*nameEntry{
		{platformID: platformMicrosoft, encodingID: encodingMicrosoftUnicode, nameID: nameFull, value: utf16BE("Solo Regular")},
		{platformID: platformMicrosoft, encodingID: encodingMicrosoftUnicode, nameID: nameFontFamily, value: utf16BE("Solo")},
	}

	font, err := newFont("solo.otf", buildFont("OTTO", entries))
	assert.NoError(t, err)
	assert.Equal(t, "Solo Regular", font.Name)
	assert.Equal(t, "Solo", font.Family)
}

func TestNewFontErrors(t *testing.T) {
	// not a font extension
	_, err := newFont("readme.txt", nil)
	assert.Error(t, err)

	// unsupported magic (woff2 is intentionally not accepted, matching the
	// extension allowlist that guards this path)
	_, err = newFont("fake.ttf", []byte("wOF2aaaaaaaaaaaaaaaa"))
	assert.ErrorIs(t, err, errUnsupportedFormat)

	// truncated data
	_, err = newFont("tiny.ttf", []byte{0x00, 0x01})
	assert.ErrorIs(t, err, errUnsupportedFormat)

	// valid header, no name table
	data := make([]byte, 12)
	copy(data, "\x00\x01\x00\x00")
	_, err = newFont("noname.ttf", data)
	assert.ErrorContains(t, err, "no name table")
}

func TestNameEntryDecoding(t *testing.T) {
	// mac roman high bytes decode via the Macintosh charmap: 0x8A is a-umlaut
	entry := &nameEntry{platformID: platformMac, encodingID: encodingMacRoman, value: []byte{'F', 0x8A}}
	assert.Equal(t, "Fä", entry.String())

	// unknown platform falls back to raw bytes
	entry = &nameEntry{platformID: 2, value: []byte("raw")}
	assert.Equal(t, "raw", entry.String())

	// unicode platform decodes UTF-16BE
	entry = &nameEntry{platformID: platformUnicode, value: utf16BE("uni")}
	assert.Equal(t, "uni", entry.String())
}
