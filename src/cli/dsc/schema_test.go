package dsc

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/invopop/jsonschema"
	basedsc "github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/stretchr/testify/assert"
)

// The embedded *.schema.json files are golden copies of what runtime
// reflection used to produce. These tests regenerate them and fail on drift;
// run with UPDATE_SCHEMAS=1 to rewrite the files after changing a state type.

func TestShellSchemaUpToDate(t *testing.T) {
	assertSchema[*Shell](t, "shell", "shell.schema.json", shellSchema)
}

func TestConfigurationSchemaUpToDate(t *testing.T) {
	assertSchema[*Configuration](t, "configuration", "configuration.schema.json", configurationSchema)
}

func assertSchema[T basedsc.State[T]](t *testing.T, id, file, embedded string) {
	t.Helper()

	generated := reflectSchema[T](id)

	if os.Getenv("UPDATE_SCHEMAS") != "" {
		assert.NoError(t, os.WriteFile(file, []byte(generated), 0644))
		return
	}

	assert.Equal(t, generated, embedded, "run UPDATE_SCHEMAS=1 go test ./cli/dsc ./cli/font to regenerate %s", file)
}

// reflectSchema mirrors the reflection the base dsc.Resource performed at
// runtime before the schemas were embedded.
func reflectSchema[T basedsc.State[T]](id string) string {
	reflector := jsonschema.Reflector{
		ExpandedStruct: true,
		DoNotReference: true,
	}

	schema := reflector.Reflect(&basedsc.Resource[T]{})
	schema.ID = jsonschema.ID(id)
	schema.Properties.Delete("$schema")
	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")

	return string(schemaJSON)
}
