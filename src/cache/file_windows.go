package cache

import (
	"bytes"
	"fmt"
	"io"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// persistentStringRWCloser implements io.ReadWriteCloser for PersistentSharedString
type persistentStringRWCloser struct {
	pss      *PersistentSharedString
	buf      *bytes.Buffer
	filePath string
	dirty    bool
}

func NewPersistentStringRWCloser(pss *PersistentSharedString) io.ReadWriteCloser {
	return &persistentStringRWCloser{
		pss:      pss,
		buf:      bytes.NewBuffer(pss.bytes()),
		filePath: pss.filePath,
	}
}

func (rw *persistentStringRWCloser) Read(p []byte) (int, error) {
	return rw.buf.Read(p)
}

func (rw *persistentStringRWCloser) Write(p []byte) (int, error) {
	if !rw.dirty {
		rw.buf.Reset()
		rw.dirty = true
	}

	return rw.buf.Write(p)
}

func (rw *persistentStringRWCloser) Close() error {
	defer rw.pss.close()

	if !rw.dirty {
		return nil
	}

	data := rw.buf.String()
	dataSize := len(data)

	// Check if the data fits in the current allocation
	if dataSize <= rw.pss.size {
		return rw.pss.SetString(data)
	}

	// Data is too large, need to recreate with larger size
	log.Debugf("cache data size (%d) exceeds current allocation (%d), recreating file", dataSize, rw.pss.size)

	// Calculate new size with some growth factor (1.5x) to reduce future reallocations
	newSize := max(dataSize+(dataSize/2), minStringSize)
	if newSize > maxStringSize {
		return fmt.Errorf("required cache size %d exceeds maximum %d", dataSize, maxStringSize)
	}

	// Close current mapping before recreating
	if err := rw.pss.close(); err != nil {
		log.Error(err)
	}

	// Create new file with larger size
	newPss, err := createOrOpenPersistentStringWithSize(rw.filePath, newSize)
	if err != nil {
		return fmt.Errorf("failed to recreate cache file with size %d: %v", newSize, err)
	}

	// Write the data to the new file
	return newPss.SetString(data)
}

func openFile(filePath string) (io.ReadWriteCloser, error) {
	pss, err := createOrOpenPersistentString(filePath)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return NewPersistentStringRWCloser(pss), nil
}
