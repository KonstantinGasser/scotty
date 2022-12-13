package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/KonstantinGasser/scotty/beam/stream"
)

func main() {

	protocol := flag.String("protocol", "unix", "protocol to stream the logs. Option: tcp")
	addr := flag.String("addr", "/tmp/scotty.sock", "address of the main process scotty")
	label := flag.String("stream", "", "labels the stream. scotty is showing information under this name (default random hash)")
	flag.Parse()

	if ok := hasPipeInput(); !ok {
		fatal("beam required you to pipe logs into beam\n\tUsage: go run -race cmd/my/program.go | beam")
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer close(sig)

	stop := make(chan struct{})
	defer close(stop)

	// propagate termination down to stream
	go func(s <-chan os.Signal, st chan<- struct{}) {
		<-s
		st <- struct{}{}
	}(sig, stop)

	conn, err := stream.Connection(*protocol, *addr)
	if err != nil {
		fatal("unable to connect to %q with protocol %q: %v", *addr, *protocol, err)
		return
	}

	if err := stream.New(*label, conn).Stream(stop); err != nil {
		fatal("unable to stream logs to scotty: %v", err)
		return
	}

}

func hasPipeInput() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false // not sure if I like this. Essentially ignoring the error..
	}

	if stat.Mode()&os.ModeCharDevice == os.ModeCharDevice || stat.Size() <= 0 {
		return false
	}

	return true
}
