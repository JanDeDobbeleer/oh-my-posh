---
id: config-sample
title: Sample
sidebar_label: Sample
---

```json
{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  "blocks": [
    {
      "alignment": "right",
      "segments": [
        {
          "foreground": "#007ACC",
          "properties": {
            "template": " {{ .CurrentDate | date .Format }} ",
            "time_format": "15:04:05"
          },
          "style": "plain",
          "type": "time"
        }
      ],
      "type": "prompt",
      "vertical_offset": -1
    },
    {
      "alignment": "left",
      "newline": true,
      "segments": [
        {
          "background": "#ffb300",
          "foreground": "#ffffff",
          "leading_diamond": "\ue0b6",
          "properties": {
            "template": " {{ .UserName }} "
          },
          "style": "diamond",
          "trailing_diamond": "\ue0b0",
          "type": "session"
        },
        {
          "background": "#61AFEF",
          "foreground": "#ffffff",
          "powerline_symbol": "\ue0b0",
          "properties": {
            "exclude_folders": [
              "/super/secret/project"
            ],
            "style": "folder",
            "template": " {{ .Path }} "
          },
          "style": "powerline",
          "type": "path"
        },
        {
          "background": "#2e9599",
          "background_templates": [
            "{{ if or (.Working.Changed) (.Staging.Changed) }}#f36943{{ end }}",
            "{{ if and (gt .Ahead 0) (gt .Behind 0) }}#a8216b{{ end }}",
            "{{ if gt .Ahead 0 }}#35b5ff{{ end }}",
            "{{ if gt .Behind 0 }}#f89cfa{{ end }}"
          ],
          "foreground": "#193549",
          "foreground_templates": [
            "{{ if and (gt .Ahead 0) (gt .Behind 0) }}#ffffff{{ end }}"
          ],
          "powerline_symbol": "\ue0b0",
          "properties": {
            "branch_max_length": 25,
            "fetch_status": true,
            "template": " {{ .HEAD }}{{ .BranchStatus }} "
          },
          "style": "powerline",
          "type": "git"
        },
        {
          "background": "#00897b",
          "background_templates": [
            "{{ if gt .Code 0 }}#e91e63{{ end }}"
          ],
          "foreground": "#ffffff",
          "properties": {
            "always_enabled": true,
            "template": "<parentBackground>\ue0b0</> \ue23a "
          },
          "style": "diamond",
          "trailing_diamond": "\ue0b4",
          "type": "exit"
        }
      ],
      "type": "prompt"
    }
  ],
  "final_space": true,
  "version": 1
}
```
