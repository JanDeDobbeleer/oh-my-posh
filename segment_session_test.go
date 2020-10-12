package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupSession(userInfoSeparator string, username string, hostname string, goos string) session {
	env := new(MockedEnvironment)
	env.On("getCurrentUser", nil).Return(username)
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
	return s
}

func testUserInfoWriter(userInfoSeparator string, username string, hostname string, goos string) string {
	s := setupSession(userInfoSeparator, username, hostname, goos)
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

func TestWriteOnlyUsername(t *testing.T) {
	s := setupSession("@", "surface\\bill", "surface", "windows")
	s.props.values[DisplayHost] = false

	want := "<#fff>bill</><#fff></>"
	got := s.getFormattedText()
	assert.EqualValues(t, want, got)
}

func TestWriteOnlyHostname(t *testing.T) {
	s := setupSession("@", "surface\\bill", "surface", "windows")
	s.props.values[DisplayUser] = false

	want := "<#fff></><#fff>surface</>"
	got := s.getFormattedText()
	assert.EqualValues(t, want, got)
}

func TestSession(t *testing.T) {
	s := &session{
		env: &environment{},
	}
	assert.NotEmpty(t, s.getUserName())
}
