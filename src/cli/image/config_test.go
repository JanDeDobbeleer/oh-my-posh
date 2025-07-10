package image

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSettings(t *testing.T) {
	cases := []struct {
		expectedResult *Settings
		name           string
		jsonContent    string
		expectError    bool
	}{
		{
			name: "Valid settings with all fields",
			jsonContent: `{
				"colors": {
					"red": "#FF0000",
					"blue": "#0000FF",
					"green": "#00FF00"
				},
				"author": "John Doe",
				"background_color": "#FFFFFF"
			}`,
			expectedResult: &Settings{
				Colors: map[string]HexColor{
					"red":   "#FF0000",
					"blue":  "#0000FF",
					"green": "#00FF00",
				},
				Author:          "John Doe",
				BackgroundColor: "#FFFFFF",
			},
			expectError: false,
		},
		{
			name: "Valid settings with only colors",
			jsonContent: `{
				"colors": {
					"red": "#FF6B6B",
					"yellow": "#FFA07A"
				}
			}`,
			expectedResult: &Settings{
				Colors: map[string]HexColor{
					"red":    "#FF6B6B",
					"yellow": "#FFA07A",
				},
				Author:          "",
				BackgroundColor: "",
			},
			expectError: false,
		},
		{
			name: "Valid settings with only author",
			jsonContent: `{
				"author": "Jane Smith"
			}`,
			expectedResult: &Settings{
				Colors:          nil,
				Author:          "Jane Smith",
				BackgroundColor: "",
			},
			expectError: false,
		},
		{
			name:        "Empty JSON object",
			jsonContent: `{}`,
			expectedResult: &Settings{
				Colors:          nil,
				Author:          "",
				BackgroundColor: "",
			},
			expectError: false,
		},
		{
			name: "Invalid JSON",
			jsonContent: `{
				"colors": {
					"red": "#FF0000"
				"author": "John Doe"
			}`,
			expectedResult: nil,
			expectError:    true,
		},
		{
			name: "JSON with invalid color format",
			jsonContent: `{
				"colors": {
					"red": "not-a-color"
				}
			}`,
			expectedResult: &Settings{
				Colors: map[string]HexColor{
					"red": "not-a-color",
				},
				Author:          "",
				BackgroundColor: "",
			},
			expectError: false,
		},
		{
			name: "JSON with extended color names",
			jsonContent: `{
				"colors": {
					"lightRed": "#FF9999",
					"darkGray": "#333333",
					"lightBlue": "#87CEEB"
				}
			}`,
			expectedResult: &Settings{
				Colors: map[string]HexColor{
					"lightRed":  "#FF9999",
					"darkGray":  "#333333",
					"lightBlue": "#87CEEB",
				},
				Author:          "",
				BackgroundColor: "",
			},
			expectError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary file
			tempFile := createTempFile(t, tc.jsonContent)
			defer os.Remove(tempFile)

			// Test LoadSettings
			result, err := LoadSettings(tempFile)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

// Helper interface for testing types that have TempDir method
type testingInterface interface {
	TempDir() string
	Helper()
}

// Helper function to create a temporary file with given content
func createTempFile(t testingInterface, content string) string {
	t.Helper()
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-settings.json")

	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		panic(err) // Use panic since we can't return error from generic interface
	}

	return tempFile
}
