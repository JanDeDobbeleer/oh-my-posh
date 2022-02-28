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
  "powerline_symbol": "\ue0b0",
  "foreground": "#ffffff",
  "background": "#8f43f3",
  "properties": {
    "precision": 2,
    "template":" \uf85a {{ round .PhysicalPercentUsed .Precision }}% "
  },
  "style":"powerline"
},
```

## Properties

- Precision: `int` - The precision used for any float values - defaults to 2

## Template ([info][templates])

:::note default template

``` template
{{ round .PhysicalPercentUsed .Precision }}
```

:::

### Properties

- `.PhysicalTotalMemory`: `int` - is the total of used physical memory
- `.PhysicalFreeMemory`: `int` - is the total of free physical memory
- `.PhysicalPercentUsed`: `float64` - is the percentage of physical memory in usage
- `.SwapTotalMemory`: `int` - is the total of used swap memory
- `.SwapFreeMemory`: `int` -  is the total of free swap memory
- `.SwapPercentUsed`: `float64` - is the percentage of swap memory in usage
- `.Load1`: `float64` - is the current load1 (can be empty on windows)
- `.Load5`: `float64` - is the current load5 (can be empty on windows)
- `.Load15`: `float64` - is the current load15 (can be empty on windows)
- `.CPU`: `[]struct` - an array of [InfoStat][cpuinfo] object, you can use any property it has e.g. `(index .CPU 0).Cores`
- `.Disks`: `[]struct` - an array of [IOCountersStat][ioinfo] object, you can use any property it has e.g. `.Disks.disk0.IoTime`

[cpuinfo]: https://github.com/shirou/gopsutil/blob/78065a7ce2021f6a78c8d6f586a2683ba501dcec/cpu/cpu.go#L32
[ioinfo]: https://github.com/shirou/gopsutil/blob/e0ec1b9cda4470db704a862282a396986d7e930c/disk/disk.go#L32
[templates]: /docs/config-templates
