package gitstatus

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

// objectStore reads commit and tree objects straight from a .git directory:
// loose objects, then pack files located via their .idx companions. It
// replaces go-git's storage/filesystem (and its dotgit/config/billy
// dependency chain) with the small subset the status engine needs. Any
// object it cannot find or decode surfaces as an error, which the caller
// turns into an exec-git fallback.
type cachedObject struct {
	kind string
	data []byte
}

// packLocation keys the per-Load object cache without allocating a string
// per lookup on the delta-chain hot path.
type packLocation struct {
	pack   *packFile
	offset int64
}

type objectStore struct {
	// cache holds inflated pack objects resolved during this Load, keyed by
	// pack offset; delta chains hit the same bases repeatedly.
	cache map[packLocation]cachedObject
	// objectsDirs is the primary objects dir plus any alternates.
	objectsDirs []string
	packs       []*packFile
	packsLoaded bool
}

func newObjectStore(commonGitDir string) *objectStore {
	primary := filepath.Join(commonGitDir, "objects")
	dirs := []string{primary}

	// objects/info/alternates lists additional object dirs, one per line
	if data, err := os.ReadFile(filepath.Join(primary, "info", "alternates")); err == nil {
		for line := range strings.SplitSeq(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if !filepath.IsAbs(line) {
				line = filepath.Join(primary, line)
			}
			dirs = append(dirs, line)
		}
	}

	return &objectStore{objectsDirs: dirs, cache: map[packLocation]cachedObject{}}
}

// close releases the pack file handles opened during this Load. Without it
// the descriptors would linger until garbage collection, which on Windows
// blocks deletion of the pack files.
func (o *objectStore) close() {
	for _, p := range o.packs {
		if !p.loaded {
			continue
		}
		p.idx.Close()
		p.pack.Close()
		p.loaded = false
		p.broken = true
	}
}

// object returns the type and inflated content of the object h.
func (o *objectStore) object(h plumbing.Hash) (kind string, data []byte, err error) {
	hex := h.String()

	for _, dir := range o.objectsDirs {
		loose := filepath.Join(dir, hex[:2], hex[2:])
		if raw, lerr := os.ReadFile(loose); lerr == nil {
			return parseLooseObject(raw)
		}
	}

	o.loadPacks()

	for _, p := range o.packs {
		offset, ok, perr := p.findOffset(h)
		if perr != nil {
			return "", nil, perr
		}
		if !ok {
			continue
		}
		return o.packObjectAt(p, offset)
	}

	return "", nil, fmt.Errorf("gitstatus: object %s not found", hex)
}

func parseLooseObject(raw []byte) (string, []byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return "", nil, err
	}
	defer zr.Close()

	content, err := io.ReadAll(zr)
	if err != nil {
		return "", nil, err
	}

	// header: "<type> <size>\x00"
	header, body, found := bytes.Cut(content, []byte{0})
	if !found {
		return "", nil, errors.New("gitstatus: malformed loose object")
	}

	kind, _, ok := strings.Cut(string(header), " ")
	if !ok {
		return "", nil, errors.New("gitstatus: malformed loose object header")
	}

	return kind, body, nil
}

func (o *objectStore) loadPacks() {
	if o.packsLoaded {
		return
	}
	o.packsLoaded = true

	for _, dir := range o.objectsDirs {
		names, err := filepath.Glob(filepath.Join(dir, "pack", "*.idx"))
		if err != nil {
			continue
		}
		for _, idxPath := range names {
			o.packs = append(o.packs, &packFile{idxPath: idxPath})
		}
	}
}

// packFile lazily opens one .idx/.pack pair. Lookups binary-search the idx
// sha table with ReadAt instead of loading the whole file: a few 20-byte
// reads per lookup even for multi-megabyte indexes.
type packFile struct {
	idx     *os.File
	pack    *os.File
	idxPath string
	fanout  [256]uint32
	loaded  bool
	broken  bool
}

const (
	idxHeaderLen = 8 // \377tOc + version
	idxFanoutLen = 256 * 4
)

func (p *packFile) open() error {
	if p.broken {
		return fmt.Errorf("gitstatus: unusable pack %s", p.idxPath)
	}
	if p.loaded {
		return nil
	}

	idx, err := os.Open(p.idxPath)
	if err != nil {
		p.broken = true
		return err
	}

	header := make([]byte, idxHeaderLen+idxFanoutLen)
	if _, err := io.ReadFull(idx, header); err != nil {
		idx.Close()
		p.broken = true
		return err
	}

	if !bytes.Equal(header[:4], []byte{0xff, 't', 'O', 'c'}) || binary.BigEndian.Uint32(header[4:8]) != 2 {
		idx.Close()
		p.broken = true
		return fmt.Errorf("gitstatus: unsupported pack index version in %s", p.idxPath)
	}

	for i := range 256 {
		p.fanout[i] = binary.BigEndian.Uint32(header[idxHeaderLen+i*4 : idxHeaderLen+i*4+4])
	}

	pack, err := os.Open(strings.TrimSuffix(p.idxPath, ".idx") + ".pack")
	if err != nil {
		idx.Close()
		p.broken = true
		return err
	}

	p.idx = idx
	p.pack = pack
	p.loaded = true
	return nil
}

// findOffset locates h in the pack index: binary search over the sorted sha
// table, then the offset tables (with 64-bit spillover for huge packs).
func (p *packFile) findOffset(h plumbing.Hash) (int64, bool, error) {
	if err := p.open(); err != nil {
		return 0, false, nil //nolint:nilerr // unusable pack: skip, other packs may hold the object
	}

	total := int64(p.fanout[255])

	lo := int64(0)
	if h[0] > 0 {
		lo = int64(p.fanout[h[0]-1])
	}
	hi := int64(p.fanout[h[0]])

	shaTableBase := int64(idxHeaderLen + idxFanoutLen)
	var sha [20]byte

	for lo < hi {
		mid := (lo + hi) / 2
		if _, err := p.idx.ReadAt(sha[:], shaTableBase+mid*20); err != nil {
			return 0, false, err
		}

		switch bytes.Compare(h[:], sha[:]) {
		case 0:
			return p.offsetAt(mid, total)
		case -1:
			hi = mid
		default:
			lo = mid + 1
		}
	}

	return 0, false, nil
}

func (p *packFile) offsetAt(pos, total int64) (int64, bool, error) {
	shaTableBase := int64(idxHeaderLen + idxFanoutLen)
	// layout after the sha table: crc32 table (4*total), 32-bit offsets
	// (4*total), then optional 64-bit offsets
	offsetBase := shaTableBase + total*20 + total*4

	var buf [8]byte
	if _, err := p.idx.ReadAt(buf[:4], offsetBase+pos*4); err != nil {
		return 0, false, err
	}

	offset32 := binary.BigEndian.Uint32(buf[:4])
	if offset32&0x80000000 == 0 {
		return int64(offset32), true, nil
	}

	largeBase := offsetBase + total*4
	largeIndex := int64(offset32 &^ 0x80000000)
	if _, err := p.idx.ReadAt(buf[:8], largeBase+largeIndex*8); err != nil {
		return 0, false, err
	}

	return int64(binary.BigEndian.Uint64(buf[:8])), true, nil
}

const (
	kindCommit = "commit"
	kindTree   = "tree"
	kindBlob   = "blob"
	kindTag    = "tag"
)

// pack object types
const (
	packCommit   = 1
	packTree     = 2
	packBlob     = 3
	packTag      = 4
	packOfsDelta = 6
	packRefDelta = 7
)

var packKinds = map[byte]string{
	packCommit: kindCommit,
	packTree:   kindTree,
	packBlob:   kindBlob,
	packTag:    kindTag,
}

// packObjectAt inflates the object at offset, resolving delta chains
// recursively.
func (o *objectStore) packObjectAt(p *packFile, offset int64) (string, []byte, error) {
	cacheKey := packLocation{pack: p, offset: offset}
	if cached, ok := o.cache[cacheKey]; ok {
		return cached.kind, cached.data, nil
	}

	header := make([]byte, 32)
	n, err := p.pack.ReadAt(header, offset)
	if err != nil && n == 0 {
		return "", nil, err
	}
	header = header[:n]

	objType := header[0] >> 4 & 7
	size := int64(header[0] & 0x0f)
	shift := uint(4)
	used := 1
	for header[used-1]&0x80 != 0 {
		if used >= len(header) {
			return "", nil, errors.New("gitstatus: pack header overflow")
		}
		size |= int64(header[used]&0x7f) << shift
		shift += 7
		used++
	}

	switch objType {
	case packOfsDelta:
		negOffset, n := decodeOffsetVarint(header[used:])
		if n == 0 {
			return "", nil, errors.New("gitstatus: malformed ofs-delta")
		}
		baseKind, base, err := o.packObjectAt(p, offset-int64(negOffset))
		if err != nil {
			return "", nil, err
		}
		return o.applyPackDelta(p, offset+int64(used+n), size, baseKind, base, cacheKey)

	case packRefDelta:
		if len(header) < used+20 {
			return "", nil, errors.New("gitstatus: malformed ref-delta")
		}
		var baseHash plumbing.Hash
		copy(baseHash[:], header[used:used+20])
		baseKind, base, err := o.object(baseHash)
		if err != nil {
			return "", nil, err
		}
		return o.applyPackDelta(p, offset+int64(used+20), size, baseKind, base, cacheKey)

	default:
		kind, ok := packKinds[objType]
		if !ok {
			return "", nil, fmt.Errorf("gitstatus: unknown pack object type %d", objType)
		}

		data, err := o.inflateAt(p, offset+int64(used), size)
		if err != nil {
			return "", nil, err
		}
		o.cache[cacheKey] = cachedObject{kind: kind, data: data}
		return kind, data, nil
	}
}

func (o *objectStore) inflateAt(p *packFile, offset, expectedSize int64) ([]byte, error) {
	section := io.NewSectionReader(p.pack, offset, 1<<40)
	zr, err := zlib.NewReader(section)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	data := make([]byte, 0, expectedSize)
	buf := bytes.NewBuffer(data)
	if _, err := io.Copy(buf, zr); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// applyPackDelta inflates the delta at deltaOffset and applies it to base.
func (o *objectStore) applyPackDelta(p *packFile, deltaOffset, deltaSize int64, baseKind string, base []byte, cacheKey packLocation) (string, []byte, error) {
	delta, err := o.inflateAt(p, deltaOffset, deltaSize)
	if err != nil {
		return "", nil, err
	}

	result, err := applyDelta(base, delta)
	if err != nil {
		return "", nil, err
	}

	o.cache[cacheKey] = cachedObject{kind: baseKind, data: result}
	return baseKind, result, nil
}

// applyDelta runs git's binary delta format: two size varints, then a
// stream of copy-from-base and insert-literal instructions.
func applyDelta(base, delta []byte) ([]byte, error) {
	baseSize, n := decodeSizeVarint(delta)
	if n == 0 || int64(len(base)) != baseSize {
		return nil, errors.New("gitstatus: delta base size mismatch")
	}
	delta = delta[n:]

	targetSize, n := decodeSizeVarint(delta)
	if n == 0 {
		return nil, errors.New("gitstatus: malformed delta")
	}
	delta = delta[n:]

	result := make([]byte, 0, targetSize)

	for len(delta) > 0 {
		op := delta[0]
		delta = delta[1:]

		if op&0x80 != 0 {
			// copy from base: bits 0-3 select offset bytes, 4-6 size bytes
			var offset, size int64
			for bit := range 4 {
				if op&(1<<bit) != 0 {
					if len(delta) == 0 {
						return nil, errors.New("gitstatus: truncated delta copy")
					}
					offset |= int64(delta[0]) << (8 * bit)
					delta = delta[1:]
				}
			}
			for bit := range 3 {
				if op&(0x10<<bit) != 0 {
					if len(delta) == 0 {
						return nil, errors.New("gitstatus: truncated delta copy")
					}
					size |= int64(delta[0]) << (8 * bit)
					delta = delta[1:]
				}
			}
			if size == 0 {
				size = 0x10000
			}
			if offset+size > int64(len(base)) {
				return nil, errors.New("gitstatus: delta copy out of range")
			}
			result = append(result, base[offset:offset+size]...)
			continue
		}

		if op == 0 {
			return nil, errors.New("gitstatus: invalid delta opcode")
		}

		// insert literal of length op
		if int(op) > len(delta) {
			return nil, errors.New("gitstatus: truncated delta insert")
		}
		result = append(result, delta[:op]...)
		delta = delta[op:]
	}

	if int64(len(result)) != targetSize {
		return nil, errors.New("gitstatus: delta target size mismatch")
	}

	return result, nil
}

// decodeSizeVarint reads git's little-endian size varint (used in delta
// headers), distinct from the big-endian offset varint.
func decodeSizeVarint(data []byte) (v int64, n int) {
	shift := uint(0)
	for {
		if n >= len(data) {
			return 0, 0
		}
		b := data[n]
		n++
		v |= int64(b&0x7f) << shift
		if b&0x80 == 0 {
			return v, n
		}
		shift += 7
	}
}
