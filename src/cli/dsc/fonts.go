package dsc

import (
	"errors"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/font"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Fonts []*Font

type Font struct {
	Name string `json:"name"`
}

func (f *Fonts) Exists(name string) bool {
	for _, font := range *f {
		if font.Name == name {
			return true
		}
	}

	return false
}

func (f *Fonts) Add(name string) {
	if font.IsLocalZipFile(name) {
		log.Debug("Skipping local zip file font:", name)
		return
	}

	if f.Exists(name) {
		log.Debug("Font already exists:", name)
		return
	}

	log.Debug("Adding font:", name)

	*f = append(*f, &Font{
		Name: name,
	})
}

func (f *Fonts) Apply(c cache.Cache) error {
	log.Debug("Applying fonts")

	font.SetCache(c)

	var err error

	for _, font := range *f {
		if installErr := font.Apply(); installErr != nil {
			log.Error(installErr)
			err = errors.Join(err, installErr)
		}
	}

	log.Debug("Fonts applied")

	return err
}

func (f *Font) Apply() error {
	asset, err := font.ResolveFontAsset(f.Name)
	if err != nil {
		return err
	}

	zipFile, err := font.Download(asset.URL)
	if err != nil {
		return err
	}

	_, err = font.InstallZIP(zipFile, asset.Folder, false)
	return err
}
