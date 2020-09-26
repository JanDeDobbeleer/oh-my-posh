package main

import (
	"encoding/json"
	"fmt"
	"os"
)

//Settings holds all the theme for rendering the prompt
type Settings struct {
	EndSpaceEnabled bool     `json:"end_space_enabled"`
	Blocks          []*Block `json:"blocks"`
}

//BlockType type of block
type BlockType string

//BlockAlignment aligment of a Block
type BlockAlignment string

const (
	//Prompt writes one or more Segments
	Prompt BlockType = "prompt"
	//LineBreak creates a line break in the prompt
	LineBreak BlockType = "newline"
	//Left aligns left
	Left BlockAlignment = "left"
	//Right aligns right
	Right BlockAlignment = "right"
)

//Block defines a part of the prompt with optional segments
type Block struct {
	Type             BlockType      `json:"type"`
	Alignment        BlockAlignment `json:"alignment"`
	HorizontalOffset int            `json:"horizontal_offset"`
	VerticalOffset   int            `json:"vertical_offset"`
	Segments         []*Segment     `json:"segments"`
}

//GetSettings returns the default configuration including possible user overrides
func GetSettings(env environmentInfo) *Settings {
	settings, err := loadUserConfiguration(env)
	if err != nil {
		return getDefaultSettings()
	}
	return settings
}

func loadUserConfiguration(env environmentInfo) (*Settings, error) {
	var settings Settings
	settingsFileLocation := fmt.Sprintf("%s/.go_my_posh", env.getenv("HOME"))
	if _, err := os.Stat(*env.getArgs().Config); !os.IsNotExist(err) {
		settingsFileLocation = *env.getArgs().Config
	}
	defaultSettings, err := os.Open(settingsFileLocation)
	defer func() {
		_ = defaultSettings.Close()
	}()
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewDecoder(defaultSettings)
	err = jsonParser.Decode(&settings)
	return &settings, err
}

func getDefaultSettings() *Settings {
	settings := &Settings{
		EndSpaceEnabled: true,
		Blocks: []*Block{
			{
				Type:      Prompt,
				Alignment: Left,
				Segments: []*Segment{
					{
						Type:            Root,
						Style:           Powerline,
						PowerlineSymbol: "",
						Background:      "#ffe9aa",
						Foreground:      "#100e23",
					},
					{
						Type:            Session,
						Style:           Powerline,
						PowerlineSymbol: "",
						Background:      "#ffffff",
						Foreground:      "#100e23",
					},
					{
						Type:            Path,
						Style:           Powerline,
						PowerlineSymbol: "",
						Background:      "#91ddff",
						Foreground:      "#100e23",
					},
					{
						Type:            Git,
						Style:           Powerline,
						PowerlineSymbol: "",
						Background:      "#95ffa4",
						Foreground:      "#100e23",
					},
					{
						Type:            Venv,
						Style:           Powerline,
						PowerlineSymbol: "",
						Background:      "#906cff",
						Foreground:      "#100e23",
					},
					{
						Type:            Exit,
						Style:           Powerline,
						PowerlineSymbol: "",
						Background:      "#ff8080",
						Foreground:      "#ffffff",
					},
				},
			},
		},
	}
	return settings
}
