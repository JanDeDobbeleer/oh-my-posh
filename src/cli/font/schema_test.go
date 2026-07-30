package font

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/invopop/jsonschema"
	basedsc "github.com/jandedobbeleer/oh-my-posh/src/dsc"
	"github.com/stretchr/testify/assert"
)

// font.schema.json is a golden copy of what runtime reflection used to
// produce; see cli/dsc/schema_test.go. Run with UPDATE_SCHEMAS=1 to rewrite.

func TestFontSchemaUpToDate(t *testing.T) {
	reflector := jsonschema.Reflector{
		ExpandedStruct: true,
		DoNotReference: true,
	}

	schema := reflector.Reflect(&basedsc.Resource[*Font]{})
	schema.ID = "font"
	schema.Properties.Delete("$schema")
	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
	generated := string(schemaJSON)

	if os.Getenv("UPDATE_SCHEMAS") != "" {
		assert.NoError(t, os.WriteFile("font.schema.json", []byte(generated), 0644))
		return
	}

	assert.Equal(t, generated, fontSchema, "run UPDATE_SCHEMAS=1 go test ./cli/font to regenerate font.schema.json")
}
