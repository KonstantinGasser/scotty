package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	protocol := flag.String("protocol", "unix", "logs can be stream/piped through unix sockets or tcp sockets")
	addr := flag.String("addr", "/tmp/scotty.sock", "specify a custom unix socket to use or a tcp:ip addr")
	label := flag.String("label", "", "logs in scotty will be displayed with the attached label")

	asDaemon := flag.Bool("d", false, "pipe logs to scotty and os.Stdout")
	flag.Parse()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	quite := make(chan struct{})
	go func(s <-chan os.Signal, q chan<- struct{}) {
		<-s
		q <- struct{}{}
	}(sig, quite)

	stream, err := newStream(*label, *protocol, *addr, *asDaemon)
	if err != nil {
		fmt.Printf("unable to open beam to scotty: %v", err)
		return
	}

	stream.beam(quite)

}
