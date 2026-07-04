//go:build !windows

package cache

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// atomicFileWriter buffers writes in memory and, on Close, writes the
// buffered content to a temporary file in the same directory and renames it
// over the destination path. os.Rename is atomic on POSIX filesystems, so
// readers either see the old, complete file or the new, complete file --
// never a torn write -- and since the temp file is created fresh, shrinking
// payloads never leave stale trailing bytes behind (unlike the previous
// O_CREATE|O_RDWR-without-O_TRUNC open, which could do both).
type atomicFileWriter struct {
	path string
	buf  bytes.Buffer
	mode os.FileMode
}

func (w *atomicFileWriter) Read(_ []byte) (int, error) {
	// Not used: store.go only reads from a freshly opened file (see
	// openFile), never from a writer being prepared for Close.
	return 0, io.EOF
}

func (w *atomicFileWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *atomicFileWriter) Close() error {
	dir := filepath.Dir(w.path)

	tmp, err := os.CreateTemp(dir, ".cache-*.tmp")
	if err != nil {
		return err
	}

	tmpName := tmp.Name()

	_, werr := tmp.Write(w.buf.Bytes())
	cerr := tmp.Close()

	if werr != nil {
		os.Remove(tmpName)
		return werr
	}

	if cerr != nil {
		os.Remove(tmpName)
		return cerr
	}

	if err := os.Chmod(tmpName, w.mode); err != nil {
		os.Remove(tmpName)
		return err
	}

	if err := os.Rename(tmpName, w.path); err != nil {
		os.Remove(tmpName)
		return err
	}

	return nil
}

// openFile opens filePath for reading the existing cache (if any). Callers
// that only need to write (store.close) should use openFileForWrite instead,
// which persists atomically.
func openFile(filePath string) (io.ReadWriteCloser, error) {
	return os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0o644)
}

// openFileForWrite returns a writer that atomically replaces filePath's
// contents on Close via a temp file + rename, preserving today's file mode.
func openFileForWrite(filePath string) (io.WriteCloser, error) {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(filePath); err == nil {
		mode = info.Mode()
	}

	return &atomicFileWriter{path: filePath, mode: mode}, nil
}
