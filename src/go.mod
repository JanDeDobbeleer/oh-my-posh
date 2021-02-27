module github.com/jandedobbeleer/oh-my-posh

go 1.16

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20201120212035-bb82daffcca2 // indirect
	github.com/distatus/battery v0.10.0
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gookit/color v1.3.8
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mitchellh/copystructure v1.1.1 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shirou/gopsutil v3.21.1+incompatible
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tklauser/go-sysconf v0.3.4 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/sys v0.0.0-20210228012217-479acdf4ea46
	golang.org/x/text v0.3.5
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	howett.net/plist v0.0.0-20201203080718-1454fab16a06 // indirect
	muzzammil.xyz/jsonc v0.0.0-20201229145248-615b0916ca38
)

replace github.com/distatus/battery v0.10.0 => github.com/JanDeDobbeleer/battery v0.10.0-1

replace github.com/gookit/color v1.3.5 => github.com/JanDeDobbeleer/color v1.3.5-1

replace github.com/shirou/gopsutil v3.21.1+incompatible => github.com/JanDeDobbeleer/gopsutil v3.21.1-1+incompatible

replace github.com/go-ole/go-ole v1.2.5 => github.com/JanDeDobbeleer/go-ole v1.2.5-1
