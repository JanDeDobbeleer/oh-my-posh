---
id: contributing_segment
title: Add Segment
sidebar_label: Add Segment
---

## Create the logic

Add a new file following this convention: `new_segment.go`.
Ensure `new` is a single verb indicating the context the segment renders.

You can use the following template as a guide.

```go
package main

type new struct {
    props          properties
    env            environmentInfo

    Text string
}

const (
    //NewProp enables something
    NewProp Property = "newprop"
)

func (n *new) enabled() bool {
    true
}

func (n *new) string() string {
    useDefaultText := n.props.getBool(NewProp, true)
    if useDefaultText {
      n.Text = "Hello"
    }
    segmentTemplate := a.props.getString(SegmentTemplate, "{{.Text}} world")
    template := &textTemplate{
      Template: segmentTemplate,
      Context:  n,
      Env:      n.env,
    }
    text, err := template.render()
    if err != nil {
      return err.Error()
    }
    return text
}

func (n *new) init(props Properties, env environmentInfo) {
    n.props = props
    n.env = env
}
```

When it comes to icon properties, make sure to use the UTF32 representation (e.g. "\uEFF1") rather than the icon itself.
This will facilitate the review process as not all environments display the icons based on the font being used.
You can find these values and query for icons easily at [Nerd Fonts][nf-icons].

For each segment, there's a single test file ensuring the functionality going forward. The convention
is `new_segment_test.go`, have a look at [existing segment tests][tests] for inspiration. Oh My Posh makes
use of the test tables pattern for all newly added tests. See [this][tables] blog post for more information.

The use of a `SegmentTemplate` is required. We're currentlly in the process of refactoring all segments to use
a template. As soon as this work is done, the templating logic will move outside of the segment's logic.

## Create a name for your Segment

[`segment.go`][segment-go] contains the list of available `SegmentType`'s, which gives them a name we can map from the
`.json` [themes][themes].

Add your segment.

```go
//New is brand new
New SegmentType = "new"
```

## Add the SegmentType mapping

Map your `SegmentType` to your Segment in the `mapSegmentWithWriter` function.

```go
New: &new{},
```

## Test your functionality

Even with unit tests, it's a good idea to build and validate the changes:

```shell
go build -o $GOPATH/bin/oh-my-posh
```

## Add the documentation

Create a new `markdown` file underneath the [`docs/docs`][docs] folder called `segment-new.md`.
Use the following template as a guide.

````markdown
---
id: new
title: New
sidebar_label: New
---

## What

Display something new.

## Sample Configuration

```json
{
  "type": "new",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#193549",
  "background": "#ffeb3b",
  "properties": {
    "newprop": "\uEFF1"
  }
}
```

## Properties

- newprop: `string` - the new text to show - defaults to `\uEFF1`
````

## Map the new documentation in the sidebar

Open [`sidebars.js`][sidebars] and add your document id (`new`) to the items of the Segments category.

## Add the JSON schema

Edit the `themes/schema.json` file to add your segment.

At `$.definitions.segment.properties.type.enum`, add your `SegmentType` to the array:

```json
new,
```

At `$.definitions.segment.allOf`, add your segment details:

```json
{
  "if": {
    "properties": {
      "type": { "const": "new" }
    }
  },
  "then": {
    "title": "Display something new",
    "description": "https://ohmyposh.dev/docs/new",
    "properties": {
      "properties": {
        "properties": {
          "nwprop": {
            "type": "boolean",
            "title": "New Property",
            "description": "the default text to display",
            "default": "Hello"
          }
        }
      }
    }
  }
}
```

## Create a pull request

And be patient, I'm going as fast as I can 🏎

[segment-go]: https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/segment.go
[themes]: https://github.com/JanDeDobbeleer/oh-my-posh/tree/main/themes
[docs]: https://github.com/JanDeDobbeleer/oh-my-posh/tree/main/docs/docs
[sidebars]: https://github.com/JanDeDobbeleer/oh-my-posh/blob/main/docs/sidebars.js
[nf-icons]: https://www.nerdfonts.com/cheat-sheet
[tests]: hhttps://github.com/JanDeDobbeleer/oh-my-posh/blob/main/src/segment_az_test.go
[tables]: https://blog.alexellis.io/golang-writing-unit-tests/
