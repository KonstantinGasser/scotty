package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	protocol := flag.String("protocol", "unix", "logs can be stream/piped through unix sockets or tcp sockets")
	addr := flag.String("addr", "/tmp/scotty.sock", "specify a custom unix socket to use or a tcp:ip addr")
	daemon := flag.Bool("d", false, "pipe logs to scotty and stdout")
	flag.Parse()

	label := flag.Arg(0)
	if len(label) <= 0 {
		printWarn("please provide a label for the stream\n\texample: \"beam engine-svc\"")
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
		printErr("unable to open beam to scotty...", err,
			"make sure scotty is running and has started with no errors",
			"if you have configured scotty to use a socket/connection different to the default,\nmake sure your beam command specifies socttys location",
		)
		return
	}

	stream.beam(quite)
}
