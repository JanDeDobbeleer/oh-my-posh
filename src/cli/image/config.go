package image

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Settings represents the structure for base 16 color overrides and other image settings.
// Expected JSON format:
//
//	{
//	  "colors": {
//	    "red": "#FF0000",
//	    "blue": "#0000FF",
//	    "green": "#00FF00"
//	  },
//	  "author": "Your Name",
//	  "background_color": "#FFFFFF"
//	}
type Settings struct {
	Colors          Colors `json:"colors"`
	Author          string `json:"author"`
	BackgroundColor string `json:"background_color"`
	Fonts           *Fonts `json:"fonts"`
	Cursor          string `json:"cursor,omitempty"`
}

type Colors map[string]HexColor

func NewColors() Colors {
	return map[string]HexColor{}
}

func LoadSettings(filePath string) (*Settings, error) {
	if filePath == "" {
		return nil, fmt.Errorf("color settings file path is empty")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read color settings file: %w", err)
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse color settings: %w", err)
	}

	return &settings, nil
}

type HexColor string

func (color HexColor) RGB() (*RGB, error) {
	hex := string(color)
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color format: %s", hex)
	}

	var r, g, b int64
	var err error

	if r, err = strconv.ParseInt(hex[0:2], 16, 64); err != nil {
		return nil, err
	}
	if g, err = strconv.ParseInt(hex[2:4], 16, 64); err != nil {
		return nil, err
	}
	if b, err = strconv.ParseInt(hex[4:6], 16, 64); err != nil {
		return nil, err
	}

	return &RGB{int(r), int(g), int(b)}, nil
}

func (colors Colors) RGBFromColorName(colorName string) (*RGB, error) {
	if colors == nil || colorName == "" {
		return nil, fmt.Errorf("colors map or colorName is empty")
	}

	if hexColor, exists := colors[colorName]; exists {
		return hexColor.RGB()
	}

	return nil, fmt.Errorf("color name '%s' not found in colors map", colorName)
}
