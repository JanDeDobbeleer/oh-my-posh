package main

import (
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testUserInfoWriter(userInfoSeparator string, username string, hostname string, goos string) string {
	env := new(MockedEnvironment)
	user := user.User{
		Username: username,
	}
	env.On("getCurrentUser", nil).Return(&user, nil)
	env.On("getHostName", nil).Return(hostname, nil)
	env.On("getRuntimeGOOS", nil).Return(goos)
	props := &properties{
		values:     map[Property]interface{}{UserInfoSeparator: userInfoSeparator},
		foreground: "#fff",
		background: "#000",
	}
	s := session{
		env:   env,
		props: props,
	}
	return s.getFormattedText()
}

func TestWriteUserInfo(t *testing.T) {
	want := "<#fff>bill</>@<#fff>surface</>"
	got := testUserInfoWriter("@", "bill", "surface", "windows")
	assert.EqualValues(t, want, got)
}

func TestWriteUserInfoWindowsIncludingHostname(t *testing.T) {
	want := "<#fff>bill</>@<#fff>surface</>"
	got := testUserInfoWriter("@", "surface\\bill", "surface", "windows")
	assert.EqualValues(t, want, got)
}
