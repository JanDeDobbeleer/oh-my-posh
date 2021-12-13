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

This example gets the latest batch information and uses the template to customize the prompt
based on the status of the batch.  The name and expected/measured Abv is displayed at all times.

If it's "Fermenting" and there are logged readings that show on Brewfather's graph, the most
recent Gravity, Temperature is shown.  Along with with an icon indicating the temperature
trend compared to the previous reading. If the most recent reading is greater than 4 hours old,
the background of the prompt is turned red - indicating an issue if, for example there is a Tilt or similar
device that is supposed to be logging to Brewfather every 15 minutes.

It will display the number of days the brew has been fermenting or conditioning as appropriate.

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
        "template":"{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d{{end}} {{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}"
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
  See the example above where I added a syringe.
  You can change the icon, put the trend elsewhere, add text, however you like!
  Make sure your NerdFont has the glyph you want or search for one
  at nerdfonts.com
- cache_timeout: `int` in minutes - How long to wait before updating the data from Brewfather.  Default is 5 minutes.

You can override the icons for temperature trend as used by template property .TemperatureTrendIcon with:

- doubleup_icon - for increases of more than 4°C, default is ↑↑
- singleup_icon - increase 2-4°C, default is ↑
- fortyfiveup_icon - increase 0.5-2°C, default is ↗
- flat_icon -change less than 0.5°C, default is →
- fortyfivedown_icon - decrease 0.5-2°C, default is ↘
- singledown_icon - decrease 2-4°C, default is ↓
- doubledown_icon - decrease more than 4°C, default is ↓↓

You can override the default icons for batch status as used by template property .StatusIcon with:

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

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[brewfather]: http://brewfather.app
