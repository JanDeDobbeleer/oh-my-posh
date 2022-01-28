---
id: sysinfo
title: System Info
sidebar_label: System Info
---

## SysInfo

Display SysInfo.

## Sample Configuration

```json
{
  "type":"sysinfo",
  "powerline_symbol": "",
  "foreground": "#ffffff",
  "background": "#8f43f3",
  "invert_powerline": true,
  "properties": {
    "prefix": " \uf85a ",
    "postfix": "% ",
    "precision": 2,
    "template":"{{ round .PhysicalPercentUsed .Precision }}"
  },
  "style":"powerline"
},
```

## Properties

- Precision: `int` - The precision used for any float values - defaults to 2

## [Template][templates] Properties

- `.PhysicalTotalMemory`: `int` - is the total of used physical memory
- `.PhysicalFreeMemory`: `int` - is the total of free physical memory
- `.PhysicalPercentUsed`: `float64` - is the percentage of physical memory in usage
- `.SwapTotalMemory`: `int` - is the total of used swap memory
- `.SwapFreeMemory`: `int` - is the percentage of swap memory in usage
- `.SwapPercentUsed`: `float64` - is the current user root/admin or not
- `.Load1`: `float64` - is the current load1 (can be empty on windows)
- `.Load5`: `float64` - is the current load5 (can be empty on windows)
- `.Load15`: `float64` - is the current load15 (can be empty on windows)
- `.CPU`: `[]struct` - an array of [InfoStat][cpuinfo] object, you can use any property it has e.g. `(index .CPU 0).Cores`

[cpuinfo]: https://github.com/shirou/gopsutil/blob/78065a7ce2021f6a78c8d6f586a2683ba501dcec/cpu/cpu.go#L32
[templates]: /docs/config-templates
