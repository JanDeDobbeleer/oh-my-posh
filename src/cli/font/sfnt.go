package font

import (
	"encoding/binary"
	"errors"
	"unicode/utf16"

	"golang.org/x/text/encoding/charmap"
)

// Minimal sfnt reader covering what font installation needs from the
// previous font-parsing dependency: the name table of a TTF/OTF file. That
// dependency also carried WOFF/WOFF2 decoding (and its brotli dependency),
// which is dead weight here - the install flow only ever feeds it .ttf and
// .otf files.

type nameID uint16

// name IDs from the OpenType name table specification
const (
	nameFontFamily      nameID = 1
	nameFull            nameID = 4
	namePreferredFamily nameID = 16
)

const (
	platformUnicode   = 0
	platformMac       = 1
	platformMicrosoft = 3

	encodingMacRoman         = 0
	encodingMicrosoftUnicode = 1
)

var errUnsupportedFormat = errors.New("unsupported font format")

type nameEntry struct {
	value      []byte
	platformID uint16
	encodingID uint16
	nameID     nameID
}

// String decodes the entry's value: UTF-16BE for Unicode and Microsoft
// Unicode entries, Mac Roman for legacy Mac entries, raw bytes for anything
// else.
func (e *nameEntry) String() string {
	unicodeEntry := e.platformID == platformUnicode ||
		(e.platformID == platformMicrosoft && e.encodingID == encodingMicrosoftUnicode)

	if unicodeEntry && len(e.value)%2 == 0 {
		u := make([]uint16, 0, len(e.value)/2)
		for i := 0; i < len(e.value); i += 2 {
			u = append(u, binary.BigEndian.Uint16(e.value[i:]))
		}

		return string(utf16.Decode(u))
	}

	if e.platformID == platformMac && e.encodingID == encodingMacRoman {
		if decoded, err := charmap.Macintosh.NewDecoder().Bytes(e.value); err == nil {
			return string(decoded)
		}
	}

	return string(e.value)
}

// readNameTable returns the raw name table of an sfnt font, or ok=false when
// the font has none.
func readNameTable(data []byte) ([]byte, bool, error) {
	if len(data) < 12 {
		return nil, false, errUnsupportedFormat
	}

	switch magic := string(data[:4]); magic {
	case "\x00\x01\x00\x00", "OTTO", "true", "typ1":
	default:
		return nil, false, errUnsupportedFormat
	}

	numTables := int(binary.BigEndian.Uint16(data[4:]))
	if len(data) < 12+numTables*16 {
		return nil, false, errUnsupportedFormat
	}

	for i := range numTables {
		record := data[12+i*16:]
		if string(record[:4]) != "name" {
			continue
		}

		offset := int64(binary.BigEndian.Uint32(record[8:]))
		length := int64(binary.BigEndian.Uint32(record[12:]))
		if offset+length > int64(len(data)) {
			return nil, false, errors.New("name table out of bounds")
		}

		return data[offset : offset+length], true, nil
	}

	return nil, false, nil
}

// parseNameTable returns the table's entries in file order, skipping records
// whose strings fall outside the table - matching how the previous parser
// surfaced them to the caller that indexes by name ID.
func parseNameTable(table []byte) ([]*nameEntry, error) {
	if len(table) < 6 {
		return nil, errors.New("name table too short")
	}

	count := int(binary.BigEndian.Uint16(table[2:]))
	stringOffset := int(binary.BigEndian.Uint16(table[4:]))

	if len(table) < 6+count*12 {
		return nil, errors.New("name table records out of bounds")
	}

	entries := make([]*nameEntry, 0, count)

	for i := range count {
		record := table[6+i*12:]
		length := int(binary.BigEndian.Uint16(record[8:]))
		offset := stringOffset + int(binary.BigEndian.Uint16(record[10:]))

		if offset+length > len(table) {
			continue
		}

		entries = append(entries, &nameEntry{
			platformID: binary.BigEndian.Uint16(record),
			encodingID: binary.BigEndian.Uint16(record[2:]),
			nameID:     nameID(binary.BigEndian.Uint16(record[6:])),
			value:      table[offset : offset+length],
		})
	}

	return entries, nil
}
