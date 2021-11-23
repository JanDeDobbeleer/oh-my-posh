---
id: nightscout
title: Nightscout
sidebar_label: Nightscout
---

## What

[Nightscout][nightscout] (CGM in the Cloud) is an open source, 
DIY project that allows real time access to a CGM data via an HTTP REST API. It 
is used for secure remote viewing of blood sugar data from anywhere...including 
OhMyPosh segments on the command line!

## Sample Configuration

This example is using mg/dl, you'll want change the numbers for mmol. Your idea 
of "high" or "low" is different from others. You'll also want to think about 
your background and foreground colors. Don't use white text on a yellow 
background, for example.

The foreground_templates example below could be set to just a single color, 
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
},
```

## Properties

- url: `string` - Your Nightscout URL, inclding the full path to entries.json 
  AND count=1 AND token. Example above. You'll know this works if you can curl 
  it yourself and get a single value. - defaults to ``
- http_timeout: `int` - How long do you want to wait before you want to see
  your prompt more than your sugar? I figure a half second is a good default - 
  defaults to 500ms
- template: `string` - a go [text/template][go-text-template] template extended 
  with [sprig][sprig] utilizing the properties below. 
  See the example above where I added a syringe. 
  You can change the icon, put the trend elsewhere, add text, however you like! 
  Make sure your NerdFont has the glyph you want or search for one 
  at nerdfonts.com 
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

## Template Properties

- .Sgv: `int` - Your Serum Glucose Value (your sugar)
- .TrendIcon: `string` - By default, this will be something like ↑↑ or ↘ etc but you can 
  override them with any glpyh as seen above

[go-text-template]: https://golang.org/pkg/text/template/
[sprig]: https://masterminds.github.io/sprig/
[nightscout]: http://www.nightscout.info/
