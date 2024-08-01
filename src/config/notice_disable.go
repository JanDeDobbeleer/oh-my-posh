//go:build disablenotice

package config

import "bytes"

func setDisableNotice(stream []byte, cfg *Config) {
	if !hasKeyInByteStream(stream, "disable_notice") {
		cfg.DisableNotice = true
	}
}

func hasKeyInByteStream(data []byte, key string) bool {
	return bytes.Contains(data, []byte(key))
}
