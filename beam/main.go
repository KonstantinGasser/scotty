package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"time"
)

func main() {

	protocol := flag.String("protocol", "unix", "protocol to stream the logs. Option: tcp")
	addr := flag.String("addr", "/tmp/scotty.sock", "address of the main process scotty")
	label := flag.String("stream", "", "labels the stream. scotty is showing information under this name (default random hash)")
	flag.Parse()

	stat, err := os.Stdin.Stat()
	if err != nil {
		fatal("unable to check stdin input: %v\n", err)
		return
	}

	if stat.Mode()&os.ModeCharDevice == os.ModeCharDevice || stat.Size() <= 0 {
		warn("Program requires input through pipes\n\tUsage: cat logs.log | beam\n")
		return
	}

	stream, err := newStream(*protocol, *label)
	if err != nil {
		fatal("unable to connect to scotty: %v\n", err)
		return
	}
	defer stream.Close()

	if err := stream.Connect(*addr); err != nil {
		fatal("unable to connect to %q: %v\n", *addr, err)
		return
	}

	stop, _ := spin(" beam me up, Scotty!")
	defer stop()

	reader := bufio.NewReader(os.Stdin)

	for {

		log, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			fatal("unable to read log: %v\n", err)
			os.Exit(1)
		}

		if _, err := stream.Write(log); err != nil {
			fatal("unable to beam log: %v\n", err)
		}
		time.Sleep(time.Second * 1)
	}

}
