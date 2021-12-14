---
id: brewfather
title: Brewfather
sidebar_label: Brewfather
---

## What

Calling all brewers!  Keep up-to-date with the status of your [Brewfather][brewfather] batch directly in your
 commandline prompt using the brewfather segment!

 You will need your User ID and API Key as generated in
 Brewfather's Settings screen, enabled with **batches.read** and **recipes.read** scopes.

## Sample Configuration

This example uses the default segment template to show a rendition of detail appropriate to the status of the batch

Additionally, the background of the segment will turn red if the latest reading is over 4 hours old - possibly helping indicate
an issue if, for example there is a Tilt or similar device that is supposed to be logging to Brewfather every 15 minutes.

NOTE: Temperature units are in degrees C and specific gravity is expressed as `1.XYZ` values.

```json
{
    "type":"brewfather",
    "style": "powerline",
    "powerline_symbol": "\uE0B0",
    "foreground": "#ffffff",
    "background": "#33158A",
    "background_templates": [
        "{{ if and (.Reading) (eq .Status \"Fermenting\") (gt .ReadingAge 4) }}#cc1515{{end}}"
    ],
    "properties": {
        "user_id":"abcdefg123456",
        "api_key":"qrstuvw78910",
        "batch_id":"hijklmno098765",
    }
},
```

## Properties

- user_id: `string` - as provided by Brewfather's Generate API Key screen.
- api_key: `string` - as provided by Brewfather's Generate API Key screen.
- batch_id: `string` - Get this by navigating to the desired batch on the brewfather website,
the batch id is at the end of the URL in the address bar.
- http_timeout: `int` in milliseconds - How long to wait for the Brewfather service to answer the request.  Default 2 seconds.
- template: `string` - a go [text/template][go-text-template] template extended
with [sprig][sprig] utilizing the properties below.
- cache_timeout: `int` in minutes - How long to wait before updating the data from Brewfather.  Default is 5 minutes.
- day_icon: `string` - icon or letter to use to indicate days.  Default is "d".

You can override the icons for temperature trend as used by template property `.TemperatureTrendIcon` with:

- doubleup_icon - for increases of more than 4°C, default is ↑↑
- singleup_icon - increase 2-4°C, default is ↑
- fortyfiveup_icon - increase 0.5-2°C, default is ↗
- flat_icon -change less than 0.5°C, default is →
- fortyfivedown_icon - decrease 0.5-2°C, default is ↘
- singledown_icon - decrease 2-4°C, default is ↓
- doubledown_icon - decrease more than 4°C, default is ↓↓

You can override the default icons for batch status as used by template property `.StatusIcon` with:

- planning_status_icon
- brewing_status_icon
- fermenting_status_icon
- conditioning_status_icon
- completed_status_icon
- archived_status_icon

## Template Properties

Commonly used fields

- .Status: `string` - One of "Planning", "Brewing", "Fermenting", "Conditioning", "Completed" or "Archived"
- .StatusIcon `string` - Icon representing above stats.  Can be overridden with properties shown above
- .TemperatureTrendIcon `string` - Icon showing temperature trend based on latest and previous reading
- .DaysFermenting `int` - days since start of fermentation
- .DaysBottled `int` - days since bottled/kegged
- .DaysBottledOrFermented `int` - one of the above, chosen automatically based on batch status
- .Recipe.Name: `string` - The recipe being brewed in this batch
- .MeasuredAbv: `float` - The ABV for the batch - either estimated from recipe or calculated from entered OG and FG values
- .ReadingAge `int` - age in hours of most recent reading or -1 if there are no readings available
  
.Reading contains the most recent data from devices or manual entry as visible on the Brewfather's batch Readings graph.
If there are no readings available, .Reading will be null.

- .Reading.Gravity: `float` - specific gravity (in decimal point format)
- .Reading.Temperature `float` - temperature in °C
- .Reading.Time `int` - unix timestamp of reading
- .Reading.Comment `string` - comment attached to this reading
- .Reading.DeviceType `string` - source of the reading, e.g. "Tilt"
- .Reading.DeviceID `string` - id of the device, e.g. "PINK"
  
Additional template properties

- .MeasuredOg: `float` - The OG for the batch as manually entered into Brewfather
- .MeasuredFg: `float` -The FG for the batch as manually entered into Brewfather
- .BrewDate: `int` - The unix timestamp of the brew day
- .FermentStartDate: `int` The unix timestamp when fermentation was started
- .BottlingDate: `time` - The unix timestamp when bottled/kegged
- .TemperatureTrend `float` - The difference between the most recent and previous temperature in °C
- .DayIcon `string` - given by "day_icon", or "d" by default

Hyperlink support

- .URL `string` - the URL for the batch in the Brewfather app.  You can use this to add a hyperlink to the segment
if you are using a terminal that supports it and the segment has `"enable_hyperlink":true` in it's properties.  `.DefaultString`
has this by default.

  Hyperlink formatting example

  ````json
  {
    // General format: [Text](Url)
    "template":"[{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d{{end}} {{.Recipe.Name}}]({{.URL}})"
  }

  ````

### Advanced Templating

The built in template will provides key useful information.  However, you can use the properties about the batch
to build your own.  For reference, the built-in template looks like this:

  ````json
  {
    "template":"{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}{{.DayIcon}} {{end}}[{{.Recipe.Name}}]({{.URL}}) {{printf \"%.1f\" .MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{printf \"%.3f\" .Reading.Gravity}} {{.Reading.Temperature}}\u00b0 {{.TemperatureTrendIcon}}{{end}}"
  }
  ````

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[brewfather]: http://brewfather.app
