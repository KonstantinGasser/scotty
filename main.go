package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/KonstantinGasser/scotty/app"
	"github.com/KonstantinGasser/scotty/store"
	"github.com/KonstantinGasser/scotty/stream"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {

	if len(os.Args) >= 2 && os.Args[1] == "version" {
		fmt.Printf("scotty:\t%s\n", version)
		return
	}

	network := flag.String("network", "unix", "network interface to listen for beams (option: tcp)")
	addr := flag.String("addr", "/tmp/scotty.sock", "address for the network interface")
	buffer := flag.Int("buffer", 4096, "buffer to store logs will hold up N items")
	refresh := flag.Duration("refresh", time.Millisecond*50, "refresh rate of the pager. Can be increased if high through put is expected in order to reduce lags")
	flag.Parse()

	quite := make(chan struct{})

	multiplex, err := stream.New(quite, *network, *addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	go multiplex.Run()

	lStore := store.New(uint32(*buffer))
	ui := app.New(quite, *refresh, lStore, multiplex)

	bubble := tea.NewProgram(ui,
		tea.WithAltScreen(),
	)

	if _, err := bubble.Run(); err != nil {
		fmt.Printf("unable to start scotty: %v", err)
		return
	}
}
