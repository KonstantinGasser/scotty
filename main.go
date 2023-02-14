package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/KonstantinGasser/scotty/app"
	"github.com/KonstantinGasser/scotty/multiplexer"
	tea "github.com/charmbracelet/bubbletea"

	"net/http"
	"net/http/pprof"
)

var version = "dev"

func main() {

	if len(os.Args) >= 1 && os.Args[1] == "version" {
		fmt.Printf("scotty:\t%s\n", version)
		return
	}

	network := flag.String("network", "unix", "network interface to listen for beams (option: tcp)")
	addr := flag.String("addr", "/tmp/scotty.sock", "address for the network interface")
	buffer := flag.Int("buffer", 4096, "buffer to store logs will hold up N items")

	quite := make(chan struct{})

	multiplex, err := multiplexer.New(quite, *network, *addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	go multiplex.Run()

	ui, err := app.New(
		*buffer,
		quite,
		multiplex.Errors(),
		multiplex.Messages(),
		multiplex.Subscribe(),
		multiplex.Unsubscribe(),
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bubble := tea.NewProgram(ui,
		tea.WithAltScreen(),
	)

	go func() {
		mux := http.NewServeMux()

		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		server := &http.Server{
			Addr:    ":8081",
			Handler: mux,
		}
		server.ListenAndServe()
	}()

	if _, err := bubble.Run(); err != nil {
		fmt.Printf("unable to start scotty: %v", err)
		return
	}
}
