package template

import (
	"strings"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestDateAcceptsUnixEpochStrings(t *testing.T) {
	previousLocal := time.Local
	time.Local = time.UTC
	defer func() {
		time.Local = previousLocal
	}()

	env := new(mock.Environment)
	env.On("Shell").Return("foo")

	Cache = new(cache.Template)
	Init(env, nil, nil)

	text, err := Render(`{{ "1776455769" | date "Mon 2/Jan 15:04" }} || {{ "1777060569" | date "Mon 2/Jan 15:04" }}`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Fri 17/Apr 19:56 || Fri 24/Apr 19:56", text)
}

func TestDateInZoneAcceptsUnixEpochStrings(t *testing.T) {
	env := new(mock.Environment)
	env.On("Shell").Return("foo")

	Cache = new(cache.Template)
	Init(env, nil, nil)

	text, err := Render(`{{ date_in_zone "Mon 2/Jan 15:04" "1776455769" "UTC" }} || {{ dateInZone "Mon 2/Jan 15:04" "1777060569" "UTC" }}`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Fri 17/Apr 19:56 || Fri 24/Apr 19:56", text)
}

func TestHTMLDateAcceptsUnixEpochStrings(t *testing.T) {
	previousLocal := time.Local
	time.Local = time.UTC
	defer func() {
		time.Local = previousLocal
	}()

	env := new(mock.Environment)
	env.On("Shell").Return("foo")

	Cache = new(cache.Template)
	Init(env, nil, nil)

	text, err := Render(`{{ htmlDate "1776455769" }} || {{ htmlDateInZone "1777060569" "UTC" }}`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "2026-04-17 || 2026-04-24", text)
}

func TestDatePipelineWithUnixEpochOutput(t *testing.T) {
	previousLocal := time.Local
	time.Local = time.UTC
	defer func() {
		time.Local = previousLocal
	}()

	env := new(mock.Environment)
	env.On("Shell").Return("foo")

	Cache = new(cache.Template)
	Init(env, nil, nil)

	text, err := Render(`{{ $n1 := (now|unixEpoch) }}{{ $n2 := (now|date_modify "+168h"|unixEpoch) }}{{ $n1 | date "2/Jan" }}||{{ $n2 | date "2/Jan" }}`, nil)
	assert.NoError(t, err)

	parts := strings.Split(text, "||")
	if assert.Len(t, parts, 2) {
		assert.NotEqual(t, parts[0], parts[1])
	}
}
