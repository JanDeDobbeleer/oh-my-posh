package shell

import "fmt"

type Features uint

const (
	Jobs Features = 1 << iota
	Azure
	PoshGit
	LineError
	Tooltips
	Transient
	FTCSMarks
	Upgrade
	Notice
	PromptMark
	RPrompt
	CursorPositioning
	Async
)

// getAllFeatures returns all defined feature flags by iterating through bit positions
func getAllFeatures() []Features {
	var features []Features

	// Iterate through all possible bit positions
	for i := range uint(32) { // 32 should be more than enough for our features
		feature := Features(1 << i)

		// Stop when we reach a power of 2 greater than our highest defined feature
		if feature > Async*2 {
			break
		}

		// Add the feature if it's a power of 2 (valid feature flag)
		if feature != 0 && (feature&(feature-1)) == 0 {
			features = append(features, feature)
		}
	}

	return features
}

func (f Features) Lines(shell string) Lines {
	var lines Lines

	// Get all features dynamically
	allFeatures := getAllFeatures()

	for _, feature := range allFeatures {
		// Check if this feature is enabled in the Features bitmask
		if uint(f)&uint(feature) == 0 {
			continue
		}

		var code Code

		switch shell {
		case PWSH:
			code = feature.Pwsh()
		case ZSH:
			code = feature.Zsh()
		case BASH:
			code = feature.Bash()
		case ELVISH:
			code = feature.Elvish()
		case FISH:
			code = feature.Fish()
		case CMD:
			code = feature.Cmd()
		case NU:
			code = feature.Nu()
		case XONSH:
			code = feature.Xonsh()
		}

		if len(code) > 0 {
			lines = append(lines, code)
		}
	}

	return lines
}

func (f Features) String() string {
	return fmt.Sprintf("%b", uint(f))
}
