package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/index"
)

func Index(directory string) (*index.Index, error) {
	r, err := git.PlainOpen(directory)
	if err != nil {
		return nil, err
	}

	return r.Storer.Index()
}
