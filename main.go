package main

import (
	"flag"
	"fmt"

	"github.com/KonstantinGasser/scotty/models/base"
	"github.com/KonstantinGasser/scotty/streams"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	network := flag.String("protocol", "unix", "type of network scotty can accept logs from")
	addr := flag.String("addr", "/tmp/scotty.sock", "address beam can connect to; beam -addr <scotty:address>")
	flag.Parse()

	multiplexer, err := streams.Open(*network, *addr)
	if err != nil {
		fmt.Printf("scotty is unable to start: %v", err)
	}

	quite := make(chan struct{}, 1)

	go multiplexer.Listen(quite)

	model := base.New(quite, multiplexer.Errors(), multiplexer.Messages())

	app := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := app.Run(); err != nil {
		fmt.Printf("unable to start scotty: %v", err)
		return
	}
}
