package font

import (
	"github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type Resource struct {
	dsc.Resource[*Font]
}

func DSC() *Resource {
	return &Resource{
		Resource: dsc.Resource[*Font]{
			JSONSchemaURL: "https://ohmyposh.dev/dsc.font.schema.json",
		},
	}
}

func (s *Resource) Apply(schema string) error {
	SetCache(s.Cache)
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
