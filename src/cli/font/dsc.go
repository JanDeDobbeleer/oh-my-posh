package font

import (
	_ "embed"

	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Resource struct {
	dsc.Resource[*Font]
}

//go:embed font.schema.json
var fontSchema string

func DSC() *Resource {
	return &Resource{
		Resource: dsc.Resource[*Font]{SchemaJSON: fontSchema},
	}
}

func (s *Resource) Apply(schema string) error {
	return s.Resource.Apply(schema)
}

func (s *Resource) Add(name string) {
	if IsLocalZipFile(name) {
		log.Debug("Skipping local zip file font:", name)
		return
	}

	s.Resource.Add(&Font{
		Name: name,
	})
}
