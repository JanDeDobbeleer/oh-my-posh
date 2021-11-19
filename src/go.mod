module oh-my-posh

go 1.17

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1
	github.com/BurntSushi/toml v0.4.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20210801044451-80ca428c5142 // indirect
	github.com/distatus/battery v0.10.0
	github.com/esimov/stackblur-go v1.0.0
	github.com/fogleman/gg v1.3.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/google/uuid v1.3.0 // indirect
	github.com/gookit/color v1.5.0
	github.com/gookit/config/v2 v2.0.27
	github.com/gookit/goutil v0.4.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.4.2
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/wayneashleyberry/terminal-dimensions v1.1.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	golang.org/x/sys v0.0.0-20211020064051-0ec99a608a1b
	golang.org/x/text v0.3.7
	gopkg.in/ini.v1 v1.63.2
	howett.net/plist v0.0.0-20201203080718-1454fab16a06 // indirect
)

require github.com/shirou/gopsutil/v3 v3.21.10

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/shopspring/decimal v1.3.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/distatus/battery v0.10.0 => github.com/JanDeDobbeleer/battery v0.10.0-2

replace github.com/shirou/gopsutil v3.21.9+incompatible => github.com/JanDeDobbeleer/gopsutil v3.21.9-1+incompatible
