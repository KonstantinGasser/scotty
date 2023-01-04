package main

import (
	"flag"
	"fmt"

	"github.com/KonstantinGasser/scotty/app"
	"github.com/KonstantinGasser/scotty/multiplexer"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	network := flag.String("network", "unix", "network interface to listen for beams (option: tcp)")
	addr := flag.String("addr", "/tmp/scotty.sock", "address for the network interface")
	flag.Parse()

	quite := make(chan struct{})

	multiplex, err := multiplexer.New(quite, *network, *addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	go multiplex.Run()

	ui, err := app.New(
		quite,
		multiplex.Errors(),
		multiplex.Messages(),
		multiplex.Beams(),
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bubble := tea.NewProgram(ui,
		tea.WithAltScreen(),
	)

	if _, err := bubble.Run(); err != nil {
		fmt.Printf("unable to start scotty: %v", err)
		return
	}
}
