package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOWDStringPlayingSong(t *testing.T) {
	expected := "滛 (16.7°)"
	d := &owm{
		temperature: 16.7,
		weather:     "滛",
	}

	assert.Equal(t, expected, d.string())
}
func bootstrapOWDTest(json string, err error) *owm {
	url := "http://api.openweathermap.org/data/2.5/weather?q=AMSTERDAM,NL&units=metric&appid=foobar"
	env := new(MockedEnvironment)
	env.On("doGet", url).Return([]byte(json), err)

	props := &properties{
		values: map[Property]interface{}{
			APIKEY:   "foobar",
			LOCATION: "AMSTERDAM,NL",
			UNITS:    "metric",
		},
	}

	owd := &owm{
		props: props,
		env:   env,
	}
	return owd
}
func TestOWDTemperatureRetrieval(t *testing.T) {
	data := `{"weather":[{"id":804,"main":"Clouds","description":"overcast clouds","icon":"04d"}],
	"main":{"temp":22.91,"feels_like":23.19,"temp_min":19.99,"temp_max":25.62,"pressure":1018,"humidity":74}}`
	d := bootstrapOWDTest(data, nil)
	err := d.setStatus()

	assert.Nil(t, err)

	assert.Equal(t, 22.91, d.temperature)
}

func TestOWDIconRetrieval(t *testing.T) {
	data := `{"weather":[{"icon":"50d"}],"main":{"temp":16}}`
	d := bootstrapOWDTest(data, nil)
	err := d.setStatus()

	assert.Nil(t, err)
	assert.Equal(t, d.weather, "")
}

func TestOWDDisabled(t *testing.T) {
	data := "nonsense"
	d := bootstrapOWDTest(data, errors.New("Whelp, something went wrong"))
	enabled := d.enabled()
	assert.False(t, enabled)
}
