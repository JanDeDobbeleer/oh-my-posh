package cli

import (
	"fmt"
	"oh-my-posh/font"

	"github.com/spf13/cobra"
)

// fontCmd can work with fonts
var fontCmd = &cobra.Command{
	Use:   "font [install|configure]",
	Short: "Manage fonts",
	Long: `Manage fonts.

This command is used to install fonts and configure the font in your terminal.

  - install: oh-my-posh font install https://github.com/ryanoasis/nerd-fonts/releases/download/v2.1.0/3270.zip`,
	ValidArgs: []string{
		"install",
		"configure",
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
		switch args[0] {
		case "install":
			font.Run()
		case "configure":
			fmt.Println("not implemented")
		default:
			_ = cmd.Help()
		}
	},
}

func init() { // nolint:gochecknoinits
	rootCmd.AddCommand(fontCmd)
}

// type fontsModel struct {
// 	fonts    []*font.Asset       // the list of choices
// 	cursor   int                 // which item our cursor is pointing at
// 	selected map[int]*font.Asset // which items are selected
// }

// func initFontsModel() (*fontsModel, error) {
// 	nerdFonts, err := font.Nerds()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &fontsModel{
// 		fonts:    nerdFonts,
// 		selected: make(map[int]*font.Asset),
// 	}, nil
// }

// func (f fontsModel) Init() tea.Cmd {
// 	// Just return `nil`, which means "no I/O right now, please."
// 	return nil
// }

// func (f fontsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	// Is it a key press?
// 	case tea.KeyMsg:
// 		// Cool, what was the actual key pressed?
// 		switch msg.String() {
// 		// These keys should exit the program.
// 		case "ctrl+c", "q":
// 			return f, tea.Quit

// 		// The "up" and "k" keys move the cursor up
// 		case "up", "k":
// 			if f.cursor > 0 {
// 				f.cursor--
// 			}

// 		// The "down" and "j" keys move the cursor down
// 		case "down", "j":
// 			if f.cursor < len(f.fonts)-1 {
// 				f.cursor++
// 			}

// 		// The "enter" key and the spacebar (a literal space) toggle
// 		// the selected state for the item that the cursor is pointing at.
// 		case "enter", " ":
// 			_, ok := f.selected[f.cursor]
// 			if ok {
// 				delete(f.selected, f.cursor)
// 			} else {
// 				f.selected[f.cursor] = f.fonts[f.cursor]
// 			}
// 		}
// 	}

// 	// Return the updated model to the Bubble Tea runtime for processing.
// 	// Note that we're not returning a command.
// 	return f, nil
// }

// func (f fontsModel) View() string {
// 	// The header
// 	s := "Which font do you want to install?\n\n"

// 	// Iterate over our choices
// 	for i, choice := range f.fonts {
// 		// Is the cursor pointing at this choice?
// 		cursor := " " // no cursor
// 		if f.cursor == i {
// 			cursor = ">" // cursor!
// 		}

// 		// Is this choice selected?
// 		checked := " " // not selected
// 		if _, ok := f.selected[i]; ok {
// 			checked = "x" // selected!
// 		}

// 		// Render the row
// 		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.Name)
// 	}

// 	// The footer
// 	s += "\nPress q to quit.\n"

// 	// Send the UI for rendering
// 	return s
// }
