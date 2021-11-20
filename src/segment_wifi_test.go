package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const netshOutput string = `
There is 1 interface on the system:

    Name                   : Wi-Fi
    Description            : Intel(R) Wireless-AC 9560 160MHz
    GUID                   : 6bb8def2-9af2-4bd4-8be2-6bd54e46bdc9
    Physical address       : d4:3b:04:e6:10:40
    State                  : connected
    SSID                   : ohsiggy
    BSSID                  : 5c:7d:7d:82:c5:73
    Network type           : Infrastructure
    Radio type             : 802.11ac
    Authentication         : WPA2-Personal
    Cipher                 : CCMP
    Connection mode        : Profile
    Channel                : 64
    Receive rate (Mbps)    : 526.5
    Transmit rate (Mbps)   : 780
    Signal                 : 94%
    Profile                : ohsiggy

    Hosted network status  : Not available


`

func bootStrapWifiTest() *wifi {
	env := new(MockedEnvironment)
	env.On("getPlatform", nil).Return(windowsPlatform)
	env.On("isWsl", nil).Return(false)
	env.On("runShellCommand", pwsh, "netsh.exe wlan show interfaces").Return(netshOutput)
	env.On("getShellName", nil).Return(pwsh)
	// env.On("hasCommand", "terraform").Return(args.hasTfCommand)
	// env.On("hasFolder", ".terraform").Return(args.hasTfFolder)
	// env.On("runCommand", "terraform", []string{"workspace", "show"}).Return(args.workspaceName, nil)
	k := &wifi{
		env:   env,
		props: &properties{},
	}
	return k
}

func TestString(t *testing.T) {
	wifi := bootStrapWifiTest()
	assert.NotNil(t, wifi.string())
}
