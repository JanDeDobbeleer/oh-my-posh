---
id: java
title: Java
sidebar_label: Java
---

## What

Display the currently active java version.

## Sample Configuration

```json
{
  "type": "java",
  "style": "powerline",
  "powerline_symbol": "\uE0B0",
  "foreground": "#ffffff",
  "background": "#4063D8",
  "template": " \uE738 {{ .Full }}"
}
```

## Properties

- home_enabled: `boolean` - display the segment in the HOME folder or not - defaults to `false`
- fetch_version: `boolean` - display the java version - defaults to `true`
- missing_command_text: `string` - text to display when the java command is missing - defaults to empty
- display_mode: `string` - determines when the segment is displayed
  - `always`: the segment is always displayed
  - `files`: the segment is only displayed when one of the following files is present:
    - `pom.xml`
    - `build.gradle.kts`
    - `build.sbt`
    - `.java-version`
    - `.deps.edn`
    - `project.clj`
    - `build.boot`
    - `*.java`
    - `*.class`
    - `*.gradle`
    - `*.jar`
    - `*.clj`
    - `*.cljc`

## Template ([info][templates])

:::note default template

``` template
{{ if .Error }}{{ .Error }}{{ else }}{{ .Full }}{{ end }}
```

:::

### Properties

- `.Full`: `string` - the full version
- `.Major`: `string` - major number
- `.Minor`: `string` - minor number
- `.Patch`: `string` - patch number
- `.Error`: `string` - error encountered when fetching the version string

[templates]: /docs/configuration/templates
