package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/KonstantinGasser/scotty/log"
	"github.com/KonstantinGasser/scotty/models/base"
	"github.com/KonstantinGasser/scotty/sock"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	network := flag.String("protocol", "unix", "type of network scotty can accept logs from")
	addr := flag.String("addr", "/tmp/scotty.sock", "address beam can connect to; beam -addr <scotty:address>")

	processor := log.NewProcessor()

	listener, err := sock.Open(*network, *addr)
	if err != nil {
		fmt.Printf("unable to start scotty; %v", err)
		return
	}

	stop := make(chan struct{}, 1)

	connC := make(chan net.Conn)
	defer close(connC)

	go sock.Listen(listener, processor, stop)

	logV := base.New(stop, processor)

	app := tea.NewProgram(logV)
	if _, err := app.Run(); err != nil {
		fmt.Printf("unable to start scotty: %v", err)
		return
	}
}
