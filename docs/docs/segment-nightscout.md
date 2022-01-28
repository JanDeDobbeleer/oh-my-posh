---
id: nightscout
title: Nightscout
sidebar_label: Nightscout
---

## What

[Nightscout][nightscout] (CGM in the Cloud) is an open source,
DIY project that allows real time access to a CGM data via an HTTP REST API. It
is used for secure remote viewing of blood sugar data from anywhere...including
Oh My Posh segments on the command line!

## Sample Configuration

This example is using mg/dl by default because the Nightscout API sends the sugar
glucose value (.Sgv) in mg/dl format. Below is also a template for displaying the
glucose value in mmol/L. When using different color ranges you should multiply your
high and low range glucose values by 18 and use these values in the templates.
You'll also want to think about your background and foreground colors. Don't use
white text on a yellow background, for example.

The `foreground_templates` example below could be set to just a single color,
if that color is visible against any of your backgrounds.

```json
{
  "type": "nightscout",
  "style": "diamond",
  "foreground": "#ffffff",
  "background": "#ff0000",
  "background_templates": [
    "{{ if gt .Sgv 150 }}#FFFF00{{ end }}",
    "{{ if lt .Sgv 60 }}#FF0000{{ end }}",
    "#00FF00"
  ],
  "foreground_templates": [
    "{{ if gt .Sgv 150 }}#000000{{ end }}",
    "{{ if lt .Sgv 60 }}#000000{{ end }}",
    "#000000"
  ],

  "leading_diamond": "",
  "trailing_diamond": "\uE0B0",
  "properties": {
    "url": "https://YOURNIGHTSCOUTAPP.herokuapp.com/api/v1/entries.json?count=1&token=APITOKENFROMYOURADMIN",
    "http_timeout": 1500,
    "template": " {{.Sgv}}{{.TrendIcon}}"
  }
}
```

Or display in mmol/l (instead of the default mg/dl) with the following template:

```json
"template": " {{ if eq (mod .Sgv 18) 0 }}{{divf .Sgv 18}}.0{{ else }} {{ round (divf .Sgv 18) 1 }}{{ end }}{{.TrendIcon}}"
```

## Properties

- url: `string` - Your Nightscout URL, including the full path to entries.json
  AND count=1 AND token. Example above. You'll know this works if you can curl
  it yourself and get a single value. - defaults to ``
- http_timeout: `int` - How long do you want to wait before you want to see
  your prompt more than your sugar? I figure a half second is a good default -
  defaults to 500ms
- NSCacheTimeout: `int` in minutes - How long do you want your numbers cached? -
  defaults to 5 min

- NOTE: You can change the icons for trend, put the trend elsewhere, add text,
  however you like!
  Make sure your NerdFont has the glyph you want or search for one at
  nerdfonts.com
- DoubleUpIcon - defaults to ↑↑
- SingleUpIcon - defaults to ↑
- FortyFiveUpIcon - defaults to ↗
- FlatIcon - defaults to →
- FortyFiveDownIcon - defaults to ↘
- SingleDownIcon - defaults to ↓
- DoubleDownIcon - defaults to ↓↓

## [Template][templates] Properties

- .ID: `string` - The internal ID of the object
- .Sgv: `int` - Your Serum Glucose Value (your sugar)
- .Date: `int` - The unix timestamp of the entry
- .DateString: `time` - The timestamp of the entry
- .Trend: `int` - The trend of the entry
- .Device: `string` - The device linked to the entry
- .Type: `string` - The type of the entry
- .UtcOffset: `int` - The UTC offset
- .SysTime: `time` - The time on the system
- .Mills: `int` - The amount of mills
- .TrendIcon: `string` - By default, this will be something like ↑↑ or ↘ etc but you can
  override them with any glpyh as seen above

[templates]: /docs/config-templates
[nightscout]: http://www.nightscout.info/
