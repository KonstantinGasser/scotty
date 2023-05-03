package debug

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func Print(msg string, args ...interface{}) {
	f, err := tea.LogToFile("scotty.log", "")
	if err != nil {
		os.Exit(1)
	}
	defer f.Close()
	fmt.Fprintf(f, msg, args...)
}
