module github.com/jandedobbeleer/oh-my-posh/src

go 1.19

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/alecthomas/assert v1.0.0
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.1.1 // indirect
	github.com/esimov/stackblur-go v1.1.0
	github.com/fogleman/gg v1.3.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/google/uuid v1.3.0 // indirect
	github.com/gookit/color v1.5.2
	github.com/gookit/config/v2 v2.1.8
	github.com/gookit/goutil v0.6.2 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/shirou/gopsutil/v3 v3.23.1
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/stretchr/testify v1.8.1
	github.com/wayneashleyberry/terminal-dimensions v1.1.0
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/image v0.3.0
	golang.org/x/sys v0.5.0
	golang.org/x/text v0.6.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/ConradIrwin/font v0.0.0-20210318200717-ce8d41cc0732
	github.com/charmbracelet/bubbles v0.15.0
	github.com/charmbracelet/bubbletea v0.23.1
	github.com/charmbracelet/lipgloss v0.6.0
	github.com/hashicorp/hcl/v2 v2.16.0
	github.com/mattn/go-runewidth v0.0.14
	github.com/spf13/cobra v1.6.1
	golang.org/x/mod v0.7.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lufia/plan9stats v0.0.0-20230110061619-bbe2e5e100de // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
)

require (
	dmitri.shuralyov.com/font/woff2 v0.0.0-20180220214647-957792cbbdab // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52 v1.2.1 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/fatih/color v1.14.1 // indirect
	github.com/goccy/go-yaml v1.9.8 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/muesli/ansi v0.0.0-20221106050444-61f0cd9a192a // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.13.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/rivo/uniseg v0.4.3 // indirect
	github.com/sahilm/fuzzy v0.1.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	github.com/zclconf/go-cty v1.12.1 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/term v0.4.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
)

replace github.com/atotto/clipboard v0.1.4 => github.com/jandedobbeleer/clipboard v0.1.4-1

replace github.com/shirou/gopsutil/v3 v3.22.10 => github.com/jandedobbeleer/gopsutil/v3 v3.22.10-1

replace github.com/gookit/config/v2 v2.1.8 => github.com/jandedobbeleer/config/v2 v2.1.8-1

replace github.com/goccy/go-yaml v1.9.8 => github.com/jandedobbeleer/go-yaml v1.9.8-2
