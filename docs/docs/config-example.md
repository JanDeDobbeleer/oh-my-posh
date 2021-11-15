---
id: config-sample
title: Sample
sidebar_label: Sample
---

```json
{
  "final_space": true,
  "blocks": [
    {
      "type": "prompt",
      "alignment": "right",
      "vertical_offset": -1,
      "segments": [
        {
          "type": "time",
          "style": "plain",
          "foreground": "#007ACC",
          "properties": {
            "time_format": "15:04:05"
          }
        }
      ]
    },
    {
      "type": "prompt",
      "alignment": "left",
      "newline": true,
      "segments": [
        {
          "type": "session",
          "style": "diamond",
          "foreground": "#ffffff",
          "background": "#ffb300",
          "leading_diamond": "\uE0B6",
          "trailing_diamond": "\uE0B0",
          "properties": {
            "postfix": " "
          }
        },
        {
          "type": "path",
          "style": "powerline",
          "powerline_symbol": "\uE0B0",
          "foreground": "#ffffff",
          "background": "#61AFEF",
          "properties": {
            "prefix": " \uE5FF ",
            "style": "folder",
            "exclude_folders": [
              "/super/secret/project"
            ],
            "enable_hyperlink": false
          }
        },
        {
          "type": "git",
          "style": "powerline",
          "foreground": "#193549",
          "foreground_templates": [
            "{{ if and (gt .Ahead 0) (gt .Behind 0) }}#ffffff{{ end }}"
          ],
          "background": "#2e9599",
          "background_templates": [
            "{{ if or (.Working.Changed) (.Staging.Changed) }}#f36943{{ end }}",
            "{{ if and (gt .Ahead 0) (gt .Behind 0) }}#a8216b{{ end }}",
            "{{ if gt .Ahead 0 }}#35b5ff{{ end }}",
            "{{ if gt .Behind 0 }}#f89cfa{{ end }}"
          ],
          "powerline_symbol": "\uE0B0",
          "properties": {
            "fetch_status": true,
            "branch_max_length": 25,
            "template": "{{ .HEAD }}{{ .BranchStatus }}"
          }
        },
        {
          "type": "exit",
          "style": "diamond",
          "foreground": "#ffffff",
          "background": "#00897b",
          "background_templates": ["{{ if gt .Code 0 }}#e91e63{{ end }}"],
          "leading_diamond": "",
          "trailing_diamond": "\uE0B4",
          "properties": {
            "always_enabled": true,
            "template": "\uE23A",
            "prefix": "<parentBackground>\uE0B0</> "
          }
        }
      ]
    }
  ]
}
```
