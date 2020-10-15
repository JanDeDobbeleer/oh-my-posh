module github.com/jandedobbeleer/oh-my-posh3

go 1.15

require (
	github.com/distatus/battery v0.10.1-0.20200722221337-7e1bf2bbb15c
	github.com/gookit/color v1.3.1
	github.com/mitchellh/go-ps v1.0.0
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/sys v0.0.0-20201015000850-e3ed0017c211
	golang.org/x/text v0.3.3
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	howett.net/plist v0.0.0-20200419221736-3b63eb3a43b5 // indirect
)

replace github.com/distatus/battery v0.10.1-0.20200722221337-7e1bf2bbb15c => github.com/JanDeDobbeleer/battery v0.10.1-0.20200909080331-bb0a7566dbb8

replace github.com/gookit/color v1.3.1 => github.com/JanDeDobbeleer/color v1.3.1-0.20201014085303-5ffcdf66388a

replace github.com/mitchellh/go-ps v1.0.0 => github.com/JanDeDobbeleer/go-ps v1.0.0

replace github.com/stretchr/testify v1.6.1 => github.com/stretchr/testify v1.6.1

replace golang.org/x/text v0.3.3 => golang.org/x/text v0.3.3
