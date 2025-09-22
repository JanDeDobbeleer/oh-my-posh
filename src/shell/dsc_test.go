package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhitespacePrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no whitespace",
			input:    "hello world",
			expected: "",
		},
		{
			name:     "bash initialization line",
			input:    `eval "$(oh-my-posh init bash)"`,
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "   ",
		},
		{
			name:     "indented bash initialization",
			input:    `    eval "$(oh-my-posh init bash)"`,
			expected: "    ",
		},
		{
			name:     "only tabs",
			input:    "\t\t\t",
			expected: "\t\t\t",
		},
		{
			name:     "tab indented initialization",
			input:    "\teval \"$(oh-my-posh init bash)\"",
			expected: "\t",
		},
		{
			name:     "mixed spaces and tabs",
			input:    " \t \t",
			expected: " \t \t",
		},
		{
			name:     "spaces before text",
			input:    "    hello world",
			expected: "    ",
		},
		{
			name:     "commented out initialization",
			input:    "# eval \"$(oh-my-posh init bash)\"",
			expected: "",
		},
		{
			name:     "tabs before text",
			input:    "\t\thello world",
			expected: "\t\t",
		},
		{
			name:     "indented comment",
			input:    "    # eval \"$(oh-my-posh init bash)\"",
			expected: "    ",
		},
		{
			name:     "mixed whitespace before text",
			input:    " \t  \teval command",
			expected: " \t  \t",
		},
		{
			name:     "powershell initialization",
			input:    "oh-my-posh init pwsh | Invoke-Expression",
			expected: "",
		},
		{
			name:     "single space",
			input:    " hello",
			expected: " ",
		},
		{
			name:     "fish initialization",
			input:    "oh-my-posh init fish | source",
			expected: "",
		},
		{
			name:     "single tab",
			input:    "\thello",
			expected: "\t",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "newline at start",
			input:    "\nhello",
			expected: "",
		},
		{
			name:     "other whitespace characters",
			input:    "\r\nhello",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := whitespacePrefix(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateShellConfig(t *testing.T) {
	tests := []struct {
		name           string
		shell          *Shell
		content        string
		expectedOutput string
		expectedUpdate bool
	}{
		{
			name: "bash - add initialization when not present",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash",
			},
			content:        "# Some existing config\necho 'hello world'\n",
			expectedOutput: "# Some existing config\necho 'hello world'\neval \"$(oh-my-posh init bash)\"\n",
			expectedUpdate: true,
		},
		{
			name: "bash - update existing initialization",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash --config theme.json",
			},
			content:        "# Some config\neval \"$(oh-my-posh init bash)\"\necho 'done'\n",
			expectedOutput: "# Some config\neval \"$(oh-my-posh init bash --config theme.json)\"\necho 'done'\n",
			expectedUpdate: true,
		},
		{
			name: "bash - no update when already correct",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash",
			},
			content:        "# Some config\neval \"$(oh-my-posh init bash)\"\necho 'done'\n",
			expectedOutput: "# Some config\neval \"$(oh-my-posh init bash)\"\necho 'done'\n",
			expectedUpdate: false,
		},
		{
			name: "zsh - add initialization when not present",
			shell: &Shell{
				Name:    "zsh",
				Command: "oh-my-posh init zsh",
			},
			content:        "# ZSH config\nexport PATH=$PATH:/usr/local/bin\n",
			expectedOutput: "# ZSH config\nexport PATH=$PATH:/usr/local/bin\neval \"$(oh-my-posh init zsh)\"\n",
			expectedUpdate: true,
		},
		{
			name: "fish - add initialization when not present",
			shell: &Shell{
				Name:    "fish",
				Command: "oh-my-posh init fish",
			},
			content:        "# Fish config\nset -x PATH $PATH /usr/local/bin\n",
			expectedOutput: "# Fish config\nset -x PATH $PATH /usr/local/bin\noh-my-posh init fish | source\n",
			expectedUpdate: true,
		},
		{
			name: "pwsh - add initialization when not present",
			shell: &Shell{
				Name:    "pwsh",
				Command: "oh-my-posh init pwsh",
			},
			content:        "# PowerShell config\n$env:PATH += ';C:\\Program Files'\n",
			expectedOutput: "# PowerShell config\n$env:PATH += ';C:\\Program Files'\noh-my-posh init pwsh | Invoke-Expression\n",
			expectedUpdate: true,
		},
		{
			name: "elvish - add initialization when not present",
			shell: &Shell{
				Name:    "elvish",
				Command: "oh-my-posh init elvish",
			},
			content:        "# Elvish config\nset paths = [$@paths /usr/local/bin]\n",
			expectedOutput: "# Elvish config\nset paths = [$@paths /usr/local/bin]\neval (oh-my-posh init elvish)\n",
			expectedUpdate: true,
		},
		{
			name: "xonsh - add initialization when not present",
			shell: &Shell{
				Name:    "xonsh",
				Command: "oh-my-posh init xonsh",
			},
			content:        "# Xonsh config\n$PATH.append('/usr/local/bin')\n",
			expectedOutput: "# Xonsh config\n$PATH.append('/usr/local/bin')\nexecx($(oh-my-posh init xonsh))\n",
			expectedUpdate: true,
		},
		{
			name: "preserve indentation when updating",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash --config new.json",
			},
			content:        "if [ -f ~/.bashrc ]; then\n    eval \"$(oh-my-posh init bash)\"\nfi\n",
			expectedOutput: "if [ -f ~/.bashrc ]; then\n    eval \"$(oh-my-posh init bash --config new.json)\"\nfi\n",
			expectedUpdate: true,
		},
		{
			name: "ignore commented oh-my-posh lines",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash",
			},
			content:        "# eval \"$(oh-my-posh init bash)\"\necho 'commented out'\n",
			expectedOutput: "# eval \"$(oh-my-posh init bash)\"\necho 'commented out'\neval \"$(oh-my-posh init bash)\"\n",
			expectedUpdate: true,
		},
		{
			name: "handle empty content",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash",
			},
			content:        "",
			expectedOutput: "\neval \"$(oh-my-posh init bash)\"\n",
			expectedUpdate: true,
		},
		{
			name: "handle content without trailing newline",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash",
			},
			content:        "echo 'no newline'",
			expectedOutput: "echo 'no newline'\neval \"$(oh-my-posh init bash)\"\n",
			expectedUpdate: true,
		},
		{
			name: "update oh-my-posh.exe reference",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash --config theme.json",
			},
			content:        "eval \"$(oh-my-posh.exe init bash)\"\n",
			expectedOutput: "eval \"$(oh-my-posh init bash --config theme.json)\"\n",
			expectedUpdate: true,
		},
		{
			name: "mixed whitespace indentation preservation",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash --config theme.json",
			},
			content:        "if command -v oh-my-posh >/dev/null 2>&1; then\n \t  eval \"$(oh-my-posh init bash)\"\nfi\n",
			expectedOutput: "if command -v oh-my-posh >/dev/null 2>&1; then\n \t  eval \"$(oh-my-posh init bash --config theme.json)\"\nfi\n",
			expectedUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, updated := tt.shell.updateShellConfig(tt.content)
			assert.Equal(t, tt.expectedOutput, result)
			assert.Equal(t, tt.expectedUpdate, updated)
		})
	}
}

func TestGetInitLinePosition(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{
			name: "find oh-my-posh init line",
			lines: []string{
				"# Some config",
				"eval \"$(oh-my-posh init bash)\"",
				"echo 'done'",
			},
			expected: 1,
		},
		{
			name: "find oh-my-posh.exe init line",
			lines: []string{
				"# Some config",
				"oh-my-posh.exe init pwsh | Invoke-Expression",
				"echo 'done'",
			},
			expected: 1,
		},
		{
			name: "ignore commented lines",
			lines: []string{
				"# eval \"$(oh-my-posh init bash)\"",
				"echo 'test'",
				"eval \"$(oh-my-posh init bash)\"",
			},
			expected: 2,
		},
		{
			name: "ignore indented commented lines",
			lines: []string{
				"    # eval \"$(oh-my-posh init bash)\"",
				"eval \"$(oh-my-posh init bash)\"",
			},
			expected: 1,
		},
		{
			name: "no oh-my-posh init line found",
			lines: []string{
				"# Some config",
				"echo 'hello'",
				"export PATH=$PATH:/usr/local/bin",
			},
			expected: -1,
		},
		{
			name: "find last occurrence",
			lines: []string{
				"eval \"$(oh-my-posh init bash)\"",
				"# another line",
				"eval \"$(oh-my-posh init bash --config theme.json)\"",
			},
			expected: 2,
		},
		{
			name: "find with extra spaces",
			lines: []string{
				"# config",
				"oh-my-posh   init   fish | source",
			},
			expected: 1,
		},
		{
			name:     "empty lines array",
			lines:    []string{},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell := &Shell{}
			result := shell.getLastInitLinePosition(tt.lines)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShellCommand(t *testing.T) {
	tests := []struct {
		name     string
		shell    *Shell
		expected string
	}{
		{
			name: "bash shell command",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash",
			},
			expected: `eval "$(oh-my-posh init bash)"`,
		},
		{
			name: "zsh shell command",
			shell: &Shell{
				Name:    "zsh",
				Command: "oh-my-posh init zsh",
			},
			expected: `eval "$(oh-my-posh init zsh)"`,
		},
		{
			name: "fish shell command",
			shell: &Shell{
				Name:    "fish",
				Command: "oh-my-posh init fish",
			},
			expected: "oh-my-posh init fish | source",
		},
		{
			name: "pwsh shell command",
			shell: &Shell{
				Name:    "pwsh",
				Command: "oh-my-posh init pwsh",
			},
			expected: "oh-my-posh init pwsh | Invoke-Expression",
		},
		{
			name: "elvish shell command",
			shell: &Shell{
				Name:    "elvish",
				Command: "oh-my-posh init elvish",
			},
			expected: `eval (oh-my-posh init elvish)`,
		},
		{
			name: "xonsh shell command",
			shell: &Shell{
				Name:    "xonsh",
				Command: "oh-my-posh init xonsh",
			},
			expected: `execx($(oh-my-posh init xonsh))`,
		},
		{
			name: "unknown shell command",
			shell: &Shell{
				Name:    "unknown",
				Command: "oh-my-posh init unknown",
			},
			expected: "oh-my-posh init unknown",
		},
		{
			name: "bash with config",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash --config theme.json",
			},
			expected: `eval "$(oh-my-posh init bash --config theme.json)"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.shell.shellCommand()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddInitLine(t *testing.T) {
	tests := []struct {
		name     string
		shell    *Shell
		content  string
		expected string
	}{
		{
			name: "add to content with trailing newline",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash",
			},
			content:  "# Some config\necho 'hello'\n",
			expected: "# Some config\necho 'hello'\neval \"$(oh-my-posh init bash)\"\n",
		},
		{
			name: "add to content without trailing newline",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash",
			},
			content:  "# Some config\necho 'hello'",
			expected: "# Some config\necho 'hello'\neval \"$(oh-my-posh init bash)\"\n",
		},
		{
			name: "add to empty content",
			shell: &Shell{
				Name:    "bash",
				Command: "oh-my-posh init bash",
			},
			content:  "",
			expected: "\neval \"$(oh-my-posh init bash)\"\n",
		},
		{
			name: "add fish init line",
			shell: &Shell{
				Name:    "fish",
				Command: "oh-my-posh init fish",
			},
			content:  "set -x PATH $PATH /usr/local/bin\n",
			expected: "set -x PATH $PATH /usr/local/bin\noh-my-posh init fish | source\n",
		},
		{
			name: "add pwsh init line",
			shell: &Shell{
				Name:    "pwsh",
				Command: "oh-my-posh init pwsh",
			},
			content:  "$env:PATH += ';C:\\\\Program Files'\n",
			expected: "$env:PATH += ';C:\\\\Program Files'\noh-my-posh init pwsh | Invoke-Expression\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.shell.addInitLine(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}
