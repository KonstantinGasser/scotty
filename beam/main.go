package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {

	protocol := flag.String("protocol", "unix", "protocol to stream the logs. Option: tcp")
	addr := flag.String("addr", "/tmp/scotty.sock", "address of the main process scotty")
	// stream := flag.String("stream", "relative path of program", "labels the stream. scotty is showing information under this name")
	flag.Parse()

	info, err := os.Stdin.Stat()
	if err != nil {
		fatal("unable to check stdin input: %v", err)
		return
	}

	if info.Mode()&os.ModeCharDevice == os.ModeCharDevice || info.Size() <= 0 {
		warn("Program requires input through pipes\n\tUsage: cat logs.log | beam")
		return
	}

	stream, err := newStream(*protocol)
	if err != nil {
		fatal("unable to connect to scotty: %v", err)
		return
	}
	defer stream.Close()

	if err := stream.Connect(*addr); err != nil {
		fatal("unable to connect to %q: %v", *addr, err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	var i = 1
	for {

		log, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		if _, err := stream.Write(log); err != nil {
			panic(err)
		}
		fmt.Printf("\rLogs send(%d)", i)
		// time.Sleep(time.Second * 2)
		i++
	}

}
