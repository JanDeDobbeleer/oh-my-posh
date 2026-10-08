// Package sqlite implements a minimal, read-only reader for the subset of
// the SQLite file format needed to look up a single row by an equality match
// on one column (e.g. gcloud's access_tokens.db). It understands table
// b-trees - interior and leaf pages, including overflow pages - well enough
// for ordinary rowid tables. It is not a general purpose SQL engine: no
// indexes, no WITHOUT ROWID tables, no query planning.
package sqlite

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const headerMagic = "SQLite format 3\x00"

const (
	pageInteriorTable = 0x05
	pageLeafTable     = 0x0d
)

// Row maps column name to its text representation. NULL values map to "".
type Row map[string]string

type database struct {
	data       []byte
	pageSize   int
	usableSize int
}

// FindRow scans table within data (a full SQLite file's bytes) for the first
// row where keyColumn equals keyValue, and returns its columns. Found is
// false when the table, or a matching row, doesn't exist.
func FindRow(data []byte, table, keyColumn, keyValue string) (row Row, found bool, err error) {
	db, err := open(data)
	if err != nil {
		return nil, false, err
	}

	rootPage, columns, err := db.tableInfo(table)
	if err != nil {
		return nil, false, err
	}

	keyIndex := -1
	for i, c := range columns {
		if strings.EqualFold(c, keyColumn) {
			keyIndex = i
			break
		}
	}

	if keyIndex < 0 {
		return nil, false, fmt.Errorf("sqlite: column %q not found in table %q", keyColumn, table)
	}

	var match Row

	walkErr := db.walkTable(rootPage, func(payload []byte) (bool, error) {
		values, err := decodeRecord(payload)
		if err != nil {
			return false, err
		}

		if keyIndex >= len(values) || values[keyIndex] != keyValue {
			return false, nil
		}

		match = make(Row, len(columns))
		for i, name := range columns {
			if i < len(values) {
				match[name] = values[i]
			}
		}

		return true, nil
	})

	if walkErr != nil {
		return nil, false, walkErr
	}

	if match == nil {
		return nil, false, nil
	}

	return match, true, nil
}

func open(data []byte) (*database, error) {
	if len(data) < 100 || string(data[:16]) != headerMagic {
		return nil, errors.New("sqlite: not a SQLite database")
	}

	pageSize := int(binary.BigEndian.Uint16(data[16:18]))
	if pageSize == 1 {
		pageSize = 65536
	}

	reserved := int(data[20])

	db := &database{
		data:       data,
		pageSize:   pageSize,
		usableSize: pageSize - reserved,
	}

	if db.pageSize < 512 || len(data) < db.pageSize {
		return nil, errors.New("sqlite: invalid page size")
	}

	return db, nil
}

func (db *database) page(n int) []byte {
	start := (n - 1) * db.pageSize
	end := start + db.pageSize

	if n < 1 || end > len(db.data) {
		return nil
	}

	return db.data[start:end]
}

// tableInfo looks up table's root page and column names from sqlite_master,
// which is always rooted at page 1.
func (db *database) tableInfo(table string) (rootPage int, columns []string, err error) {
	found := false

	walkErr := db.walkTable(1, func(payload []byte) (bool, error) {
		values, err := decodeRecord(payload)
		if err != nil {
			return false, err
		}

		// sqlite_master columns: type, name, tbl_name, rootpage, sql
		if len(values) < 5 || values[0] != "table" || values[2] != table {
			return false, nil
		}

		rp, err := strconv.Atoi(values[3])
		if err != nil {
			return false, fmt.Errorf("sqlite: invalid rootpage for table %q: %w", table, err)
		}

		cols, err := parseColumns(values[4])
		if err != nil {
			return false, err
		}

		rootPage, columns, found = rp, cols, true

		return true, nil
	})

	if walkErr != nil {
		return 0, nil, walkErr
	}

	if !found {
		return 0, nil, fmt.Errorf("sqlite: table %q not found", table)
	}

	return rootPage, columns, nil
}

// parseColumns extracts column names, in declaration order, from a CREATE
// TABLE statement. It only handles the simple, single-line column lists
// gcloud uses, skipping table-level constraints (PRIMARY KEY, UNIQUE, etc.)
// that don't correspond to a stored column.
func parseColumns(sql string) ([]string, error) {
	start := strings.IndexByte(sql, '(')
	end := strings.LastIndexByte(sql, ')')

	if start < 0 || end < 0 || end <= start {
		return nil, fmt.Errorf("sqlite: malformed CREATE TABLE: %s", sql)
	}

	body := sql[start+1 : end]

	var fields []string
	depth := 0
	last := 0

	for i, r := range body {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				fields = append(fields, body[last:i])
				last = i + 1
			}
		}
	}

	fields = append(fields, body[last:])

	tableConstraints := map[string]bool{
		"PRIMARY": true, "UNIQUE": true, "CHECK": true, "FOREIGN": true, "CONSTRAINT": true,
	}

	var columns []string

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		name := firstToken(field)
		if tableConstraints[strings.ToUpper(name)] {
			continue
		}

		columns = append(columns, trimIdentifier(name))
	}

	return columns, nil
}

func firstToken(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	switch s[0] {
	case '"', '`', '\'':
		if end := strings.IndexByte(s[1:], s[0]); end >= 0 {
			return s[:end+2]
		}
	case '[':
		if end := strings.IndexByte(s, ']'); end >= 0 {
			return s[:end+1]
		}
	}

	if i := strings.IndexAny(s, " \t\r\n("); i >= 0 {
		return s[:i]
	}

	return s
}

func trimIdentifier(s string) string {
	if len(s) < 2 {
		return s
	}

	switch s[0] {
	case '"', '`', '\'':
		if s[len(s)-1] == s[0] {
			return s[1 : len(s)-1]
		}
	case '[':
		if s[len(s)-1] == ']' {
			return s[1 : len(s)-1]
		}
	}

	return s
}

// walkTable visits every row's record payload, in storage order, in the
// table b-tree rooted at rootPage. visit returns (stop, err); walking ends
// early when stop is true or err is non-nil.
func (db *database) walkTable(rootPage int, visit func(payload []byte) (bool, error)) error {
	_, err := db.walkPage(rootPage, visit)
	return err
}

func (db *database) walkPage(pageNum int, visit func(payload []byte) (bool, error)) (bool, error) {
	page := db.page(pageNum)
	if page == nil {
		return false, fmt.Errorf("sqlite: page %d out of range", pageNum)
	}

	headerOffset := 0
	if pageNum == 1 {
		headerOffset = 100
	}

	pageType := page[headerOffset]

	var headerSize int
	switch pageType {
	case pageLeafTable:
		headerSize = 8
	case pageInteriorTable:
		headerSize = 12
	default:
		return false, fmt.Errorf("sqlite: unsupported page type 0x%02x", pageType)
	}

	numCells := int(binary.BigEndian.Uint16(page[headerOffset+3 : headerOffset+5]))
	cellPointerStart := headerOffset + headerSize

	for i := range numCells {
		offsetPos := cellPointerStart + i*2
		cellOffset := int(binary.BigEndian.Uint16(page[offsetPos : offsetPos+2]))

		if pageType == pageInteriorTable {
			childPage := int(binary.BigEndian.Uint32(page[cellOffset : cellOffset+4]))

			stop, err := db.walkPage(childPage, visit)
			if err != nil || stop {
				return stop, err
			}

			continue
		}

		payload, err := db.readLeafCell(page, cellOffset)
		if err != nil {
			return false, err
		}

		stop, err := visit(payload)
		if err != nil || stop {
			return stop, err
		}
	}

	if pageType == pageInteriorTable {
		rightMost := int(binary.BigEndian.Uint32(page[headerOffset+8 : headerOffset+12]))
		return db.walkPage(rightMost, visit)
	}

	return false, nil
}

// readLeafCell decodes a table leaf cell's record payload, reassembling it
// from overflow pages when the payload doesn't fit on the page.
func (db *database) readLeafCell(page []byte, offset int) ([]byte, error) {
	payloadLen, n := readVarint(page[offset:])
	offset += n

	_, n = readVarint(page[offset:]) // rowid, unused: we filter by column value
	offset += n

	total := int(payloadLen)

	maxLocal := db.usableSize - 35
	if total <= maxLocal {
		if offset+total > len(page) {
			return nil, errors.New("sqlite: cell payload exceeds page bounds")
		}

		return page[offset : offset+total], nil
	}

	minLocal := (db.usableSize-12)*32/255 - 23
	local := minLocal + (total-minLocal)%(db.usableSize-4)

	if local > maxLocal {
		local = minLocal
	}

	payload := make([]byte, 0, total)
	payload = append(payload, page[offset:offset+local]...)

	overflowPage := int(binary.BigEndian.Uint32(page[offset+local : offset+local+4]))
	remaining := total - local

	for overflowPage != 0 && remaining > 0 {
		op := db.page(overflowPage)
		if op == nil {
			return nil, fmt.Errorf("sqlite: overflow page %d out of range", overflowPage)
		}

		next := int(binary.BigEndian.Uint32(op[0:4]))

		chunk := min(db.usableSize-4, remaining)

		payload = append(payload, op[4:4+chunk]...)
		remaining -= chunk
		overflowPage = next
	}

	return payload, nil
}

// readVarint reads a SQLite variable-length integer from the start of buf,
// returning its value and the number of bytes consumed (1-9).
func readVarint(buf []byte) (int64, int) {
	var v int64

	for i := range min(8, len(buf)) {
		b := buf[i]
		v = (v << 7) | int64(b&0x7f)

		if b&0x80 == 0 {
			return v, i + 1
		}
	}

	if len(buf) < 9 {
		return v, len(buf)
	}

	v = (v << 8) | int64(buf[8])

	return v, 9
}

// decodeRecord decodes a SQLite record's column values into their text
// representation, in column order.
func decodeRecord(payload []byte) ([]string, error) {
	headerLen, n := readVarint(payload)
	if headerLen <= 0 || int(headerLen) > len(payload) {
		return nil, errors.New("sqlite: invalid record header length")
	}

	headerEnd := int(headerLen)
	pos := n

	var serialTypes []int64

	for pos < headerEnd {
		st, used := readVarint(payload[pos:])
		if used == 0 {
			return nil, errors.New("sqlite: truncated record header")
		}

		serialTypes = append(serialTypes, st)
		pos += used
	}

	values := make([]string, len(serialTypes))
	body := payload[headerEnd:]
	offset := 0

	for i, st := range serialTypes {
		size := serialTypeSize(st)
		if offset+size > len(body) {
			return nil, errors.New("sqlite: record body shorter than declared")
		}

		values[i] = decodeValue(st, body[offset:offset+size])
		offset += size
	}

	return values, nil
}

func serialTypeSize(st int64) int {
	switch {
	case st == 0, st == 8, st == 9:
		return 0
	case st >= 1 && st <= 4:
		return int(st)
	case st == 5:
		return 6
	case st == 6, st == 7:
		return 8
	case st >= 12 && st%2 == 0:
		return int((st - 12) / 2)
	case st >= 13:
		return int((st - 13) / 2)
	default:
		return 0
	}
}

func decodeValue(st int64, b []byte) string {
	switch {
	case st == 0:
		return ""
	case st == 8:
		return "0"
	case st == 9:
		return "1"
	case st >= 1 && st <= 6:
		return strconv.FormatInt(decodeBigEndianInt(b), 10)
	case st == 7:
		return strconv.FormatFloat(math.Float64frombits(binary.BigEndian.Uint64(b)), 'g', -1, 64)
	default:
		// blob (even, >=12) or text (odd, >=13): both carry raw bytes.
		return string(b)
	}
}

// decodeBigEndianInt decodes a sign-extended, big-endian two's complement
// integer of 1, 2, 3, 4, 6 or 8 bytes, as used by SQLite record serial types 1-6.
func decodeBigEndianInt(b []byte) int64 {
	var v int64

	for i, c := range b {
		if i == 0 && c&0x80 != 0 {
			v = -1
		}

		v = (v << 8) | int64(c)
	}

	return v
}
