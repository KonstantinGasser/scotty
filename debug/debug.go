package debug

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// dev
func Debug(msg string) {
	f, err := tea.LogToFile("scotty.log", "")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	log.Printf("%q", msg)
	defer f.Close()
}
