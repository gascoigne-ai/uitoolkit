package sections_test

import (
	"fmt"

	"github.com/gascoigne-ai/uitoolkit/examples/showcase/sections"
)

// The nav lists the sections in the order of the table.
func Example_sections() {
	for _, section := range sections.Sections {
		fmt.Println(section.Label)
	}
	// Output:
	// Buttons and chips
	// Text input
	// Toggles and sliders
	// Lists and panes
	// Menus and dialogs
	// Code view
	// Images and icons
	// Typography
	// Theming
}
