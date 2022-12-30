package main

import (
	"fmt"

	"github.com/KonstantinGasser/scotty/app"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	ui, err := app.New()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bubble := tea.NewProgram(ui,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := bubble.Run(); err != nil {
		fmt.Printf("unable to start scotty: %v", err)
		return
	}
}
