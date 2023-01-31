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

func Print(msg string, args ...interface{}) {
	f, err := tea.LogToFile("scotty.log", "")
	if err != nil {
		os.Exit(1)
	}
	defer f.Close()
	defer fmt.Fprintln(f, "+++++++++++++++++++++++++++++++++++++")
	fmt.Fprintln(f, "+++++++++++++++++++++++++++++++++++++")
	fmt.Fprintf(f, msg, args...)
}
