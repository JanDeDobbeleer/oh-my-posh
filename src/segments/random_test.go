package segments

import (
	"slices"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestRandomSegmentTemplate(t *testing.T) {
	cases := []struct {
		Case          string
		RandomOptions interface{}
		ExpectedIn    []string
	}{
		{
			Case:          "default options",
			RandomOptions: nil,
			ExpectedIn:    []string{"a", "b", "c"},
		},
		{
			Case:          "custom options",
			RandomOptions: []string{"x", "y", "z"},
			ExpectedIn:    []string{"x", "y", "z"},
		},
		{
			Case:          "single option",
			RandomOptions: []string{"x"},
			ExpectedIn:    []string{"x"},
		},
		{
			Case:          "empty string option",
			RandomOptions: []string{""},
			ExpectedIn:    []string{""},
		},
		{
			Case:          "no options",
			RandomOptions: []string{},
			ExpectedIn:    []string{"a", "b", "c"},
		},
		{
			Case:          "emojis",
			RandomOptions: []string{"👍", "👎", "🤷"},
			ExpectedIn:    []string{"👍", "👎", "🤷"},
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		random := &Random{}
		props := properties.Map{
			RandomOptions: tc.RandomOptions,
		}
		random.Init(props, env)

		assert.Equal(t, true, random.Enabled())
		assert.Contains(t, tc.ExpectedIn, renderTemplate(env, random.Template(), random), tc.Case)
		if len(tc.ExpectedIn) > 1 {
			assert.Equal(t, true, checkIsRandom(env, random), tc.Case)
		}
	}
}

func checkIsRandom(env *mock.Environment, random *Random) bool {
	results := []string{random.Text}
	for i := 0; i < 20; i++ {
		random.Enabled()
		text := renderTemplate(env, random.Template(), random)

		if !slices.Contains(results, text) {
			// Found a new value
			return true
		}
		results = append(results, text)
	}
	return false
}
