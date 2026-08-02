package gitstatus

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
)

// Minimal git index parser, replacing go-git's plumbing/format/index decoder.
// It decodes only what the status engine needs, skips the trailing checksum
// (which also transparently supports index.skipHash), and allocates one
// entry struct plus one name string per entry.
//
// Format reference: Documentation/gitformat-index.txt in git itself.

const (
	indexEntryHeaderLen = 62 // fixed-width part of an on-disk entry, v2/v3

	// mode values as stored on disk (octal)
	modeSymlink = 0o120000
	modeGitlink = 0o160000

	flagExtended     = 0x4000
	flagStageMask    = 0x3000
	flagNameMask     = 0x0fff
	flagSkipWorktree = 1 << 14 // in the v3 extended flags word
	flagIntentToAdd  = 1 << 13
)

var (
	errIndexMalformed = errors.New("gitstatus: malformed index file")

	indexSignature = [4]byte{'D', 'I', 'R', 'C'}
)

// indexEntry is the subset of an index entry the status engine uses.
type indexEntry struct {
	Name         string
	Hash         plumbing.Hash
	MtimeSec     uint32
	MtimeNsec    uint32
	Size         uint32
	Mode         uint32
	Stage        uint8
	SkipWorktree bool
	IntentToAdd  bool
}

// cacheTree is one entry of the TREE extension. Entries < 0 marks the span
// invalidated.
type cacheTree struct {
	Path    string
	Hash    plumbing.Hash
	Entries int
}

type gitIndex struct {
	Entries   []indexEntry
	CacheTree []cacheTree
}

// decodeIndex parses an index file. It fails on unsupported versions and on
// mandatory extensions it does not understand (sparse index `sdir`, split
// index `link`), which is exactly the signal the caller uses to fall back to
// exec git.
func decodeIndex(data []byte) (*gitIndex, error) {
	if len(data) < 12 || [4]byte(data[:4]) != indexSignature {
		return nil, errIndexMalformed
	}

	version := binary.BigEndian.Uint32(data[4:8])
	if version < 2 || version > 4 {
		return nil, fmt.Errorf("gitstatus: unsupported index version %d", version)
	}

	count := binary.BigEndian.Uint32(data[8:12])
	idx := &gitIndex{Entries: make([]indexEntry, 0, count)}

	rest, err := decodeEntries(data[12:], version, count, idx)
	if err != nil {
		return nil, err
	}

	return idx, decodeExtensions(rest, idx)
}

func decodeEntries(data []byte, version, count uint32, idx *gitIndex) ([]byte, error) {
	var prevName []byte

	for range count {
		if len(data) < indexEntryHeaderLen {
			return nil, errIndexMalformed
		}

		e := indexEntry{
			MtimeSec:  binary.BigEndian.Uint32(data[8:12]),
			MtimeNsec: binary.BigEndian.Uint32(data[12:16]),
			Mode:      binary.BigEndian.Uint32(data[24:28]),
			Size:      binary.BigEndian.Uint32(data[36:40]),
		}
		copy(e.Hash[:], data[40:60])

		flags := binary.BigEndian.Uint16(data[60:62])
		e.Stage = uint8(flags & flagStageMask >> 12)
		nameLen := int(flags & flagNameMask)

		consumed := indexEntryHeaderLen
		if flags&flagExtended != 0 {
			if version < 3 || len(data) < consumed+2 {
				return nil, errIndexMalformed
			}
			extended := binary.BigEndian.Uint16(data[consumed : consumed+2])
			e.SkipWorktree = extended&flagSkipWorktree != 0
			e.IntentToAdd = extended&flagIntentToAdd != 0
			consumed += 2
		}

		var name []byte
		var err error
		if version == 4 {
			name, data, err = decodeNameV4(data[consumed:], prevName)
			if err != nil {
				return nil, err
			}
		} else {
			name, data, err = decodeNamePadded(data, consumed, nameLen)
			if err != nil {
				return nil, err
			}
		}

		prevName = name
		e.Name = string(name)
		idx.Entries = append(idx.Entries, e)
	}

	return data, nil
}

// decodeNamePadded reads a v2/v3 entry name and skips the NUL padding that
// aligns the entry to a multiple of 8 bytes.
func decodeNamePadded(entry []byte, consumed, nameLen int) (name, rest []byte, err error) {
	if nameLen == flagNameMask {
		// name is 0xfff or longer: stored NUL-terminated instead
		nul := -1
		for i := consumed; i < len(entry); i++ {
			if entry[i] == 0 {
				nul = i
				break
			}
		}
		if nul < 0 {
			return nil, nil, errIndexMalformed
		}
		nameLen = nul - consumed
	}

	if len(entry) < consumed+nameLen {
		return nil, nil, errIndexMalformed
	}
	name = entry[consumed : consumed+nameLen]

	// pad the whole entry (header + name + at least one NUL) to 8 bytes
	entryLen := (consumed + nameLen + 8) / 8 * 8
	if len(entry) < entryLen {
		return nil, nil, errIndexMalformed
	}

	return name, entry[entryLen:], nil
}

// decodeNameV4 reads a version-4 prefix-compressed name: a varint counting
// bytes to strip from the previous name, then a NUL-terminated suffix.
func decodeNameV4(data, prevName []byte) (name, rest []byte, err error) {
	strip, n := decodeOffsetVarint(data)
	if n == 0 || int(strip) > len(prevName) {
		return nil, nil, errIndexMalformed
	}
	data = data[n:]

	nul := -1
	for i := range data {
		if data[i] == 0 {
			nul = i
			break
		}
	}
	if nul < 0 {
		return nil, nil, errIndexMalformed
	}

	keep := prevName[:len(prevName)-int(strip)]
	name = make([]byte, 0, len(keep)+nul)
	name = append(name, keep...)
	name = append(name, data[:nul]...)

	return name, data[nul+1:], nil
}

// decodeOffsetVarint reads git's offset-encoded variable-width integer (the
// same encoding packfiles use for delta base offsets).
func decodeOffsetVarint(data []byte) (v uint64, n int) {
	if len(data) == 0 {
		return 0, 0
	}

	b := data[0]
	n = 1
	v = uint64(b & 0x7f)

	for b&0x80 != 0 {
		if n >= len(data) {
			return 0, 0
		}
		b = data[n]
		n++
		v = (v+1)<<7 | uint64(b&0x7f)
	}

	return v, n
}

// decodeExtensions walks the extension blocks after the entries. Optional
// extensions (signature starting with A-Z) other than TREE are skipped;
// mandatory ones are a hard error so the caller falls back to exec git.
// The trailing checksum is not verified: git itself makes it optional via
// index.skipHash.
func decodeExtensions(data []byte, idx *gitIndex) error {
	const trailerLen = 20

	for len(data) > trailerLen {
		if len(data) < 8+trailerLen {
			return errIndexMalformed
		}

		signature := string(data[:4])
		size := int(binary.BigEndian.Uint32(data[4:8]))
		if len(data) < 8+size+trailerLen {
			return errIndexMalformed
		}

		payload := data[8 : 8+size]
		switch {
		case signature == "TREE":
			if err := decodeTreeExtension(payload, idx); err != nil {
				return err
			}
		case signature[0] < 'A' || signature[0] > 'Z':
			return fmt.Errorf("gitstatus: mandatory index extension %q not supported", signature)
		}

		data = data[8+size:]
	}

	return nil
}

// decodeTreeExtension parses the cache-tree extension: a sequence of
// "path NUL entry-count SP subtree-count LF [hash]" records, hash present
// only for valid (non-negative entry-count) records.
func decodeTreeExtension(data []byte, idx *gitIndex) error {
	for len(data) > 0 {
		nul := -1
		for i := range data {
			if data[i] == 0 {
				nul = i
				break
			}
		}
		if nul < 0 {
			return errIndexMalformed
		}

		entry := cacheTree{Path: string(data[:nul])}
		data = data[nul+1:]

		count, rest, err := readASCIIInt(data, ' ')
		if err != nil {
			return err
		}
		entry.Entries = count

		_, rest, err = readASCIIInt(rest, '\n')
		if err != nil {
			return err
		}
		data = rest

		if count >= 0 {
			if len(data) < 20 {
				return errIndexMalformed
			}
			copy(entry.Hash[:], data[:20])
			data = data[20:]
		}

		idx.CacheTree = append(idx.CacheTree, entry)
	}

	return nil
}

func readASCIIInt(data []byte, delim byte) (int, []byte, error) {
	end := -1
	for i := range data {
		if data[i] == delim {
			end = i
			break
		}
	}
	if end < 0 {
		return 0, nil, errIndexMalformed
	}

	v := 0
	neg := false
	for i, c := range data[:end] {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return 0, nil, errIndexMalformed
		}
		v = v*10 + int(c-'0')
	}
	if neg {
		v = -v
	}

	return v, data[end:][1:], nil
}
