package maps

import (
	"encoding/gob"
)

func init() {
	gob.Register(&Config{})
	gob.Register(&Map{})
}

type Config struct {
	UserName  *Map `json:"user_name,omitempty" toml:"user_name,omitempty" yaml:"user_name,omitempty"`
	HostName  *Map `json:"host_name,omitempty" toml:"host_name,omitempty" yaml:"host_name,omitempty"`
	ShellName *Map `json:"shell_name,omitempty" toml:"shell_name,omitempty" yaml:"shell_name,omitempty"`
}

func (c *Config) GetUserName(key string) string {
	if c == nil || c.UserName == nil {
		return key
	}

	return c.UserName.Get(key)
}

func (c *Config) GetHostName(key string) string {
	if c == nil || c.HostName == nil {
		return key
	}

	return c.HostName.Get(key)
}

func (c *Config) GetShellName(key string) string {
	if c == nil || c.ShellName == nil {
		return key
	}

	return c.ShellName.Get(key)
}

type Map map[string]string

func (m *Map) Get(key string) string {
	if m == nil {
		return key
	}

	if value, ok := (*m)[key]; ok {
		return value
	}

	return key
}
