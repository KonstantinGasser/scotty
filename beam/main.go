package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
)

func main() {

	protocol := flag.String("protocol", "unix", "logs can be stream/piped through unix sockets or tcp sockets")
	addr := flag.String("addr", "/tmp/scotty.sock", "specify a custom unix socket to use or a tcp:ip addr")
	daemon := flag.Bool("d", false, "pipe logs to scotty and os.Stdout")
	flag.Parse()

	label := flag.Arg(0)
	if len(label) <= 0 {
		fmt.Println(
			lipgloss.NewStyle().Foreground(
				lipgloss.Color("#ff0000"),
			).Render(
				"please provide a label for the stream\n\texample: \"beam engine-svc\"",
			),
		)
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	quite := make(chan struct{})
	go func(s <-chan os.Signal, q chan<- struct{}) {
		<-s
		q <- struct{}{}
	}(sig, quite)

	stream, err := newStream(label, *protocol, *addr, *daemon)
	if err != nil {
		fmt.Println(
			lipgloss.NewStyle().Foreground(
				lipgloss.Color("#ff0000"),
			).Render(
				fmt.Sprintf("unable to open beam to scotty: %v", err),
			),
		)
		return
	}

	stream.beam(quite)
}
