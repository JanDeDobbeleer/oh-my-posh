package cache

import (
	"bytes"
	"io"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// persistentStringRWCloser implements io.ReadWriteCloser for PersistentSharedString
type persistentStringRWCloser struct {
	pss   *PersistentSharedString
	buf   *bytes.Buffer
	dirty bool
}

func NewPersistentStringRWCloser(pss *PersistentSharedString) io.ReadWriteCloser {
	return &persistentStringRWCloser{
		pss: pss,
		buf: bytes.NewBuffer(pss.bytes()),
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

	if rw.dirty {
		return rw.pss.SetString(rw.buf.String())
	}

	return nil
}

func openFile(filePath string) (io.ReadWriteCloser, error) {
	pss, err := createOrOpenPersistentString(filePath)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return NewPersistentStringRWCloser(pss), nil
}
