package main

import (
	"flag"
	"fmt"

	"github.com/KonstantinGasser/scotty/streams"
	"github.com/KonstantinGasser/scotty/ui"
	"github.com/KonstantinGasser/scotty/ui/common"
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

	// model, err := base.New(quite,
	// 	multiplexer.Errors(),
	// 	multiplexer.Subscribers(),
	// 	multiplexer.Messages(),
	// )
	// if err != nil {
	// 	fmt.Printf("unable to start scotty: %v", err)
	// 	return
	// }

	width, height, err := common.WindowSize()

	if err != nil {
		fmt.Printf("unable to determine terminal width and height: %v", err)
		return
	}
	model := ui.New(width, height, quite)
	app := tea.NewProgram(model, tea.WithAltScreen(), tea.WithoutCatchPanics())
	if _, err := app.Run(); err != nil {
		fmt.Printf("unable to start scotty: %v", err)
		return
	}
}
