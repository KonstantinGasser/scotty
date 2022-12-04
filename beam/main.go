package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {

	socket := flag.String("socket", "unix:/temp", "connection to stream logs over. Use ip:port to connect to a tcp:ip socket")
	// stream := flag.String("stream", "relative path of programm", "labels the stream. scotty is showing information under this name")
	flag.Parse()

	sock, err := newUnix(*socket)
	if err != nil {
		panic(err)
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

		if _, err := sock.Write(log); err != nil {
			panic(err)
		}
		fmt.Printf("\rLogs send(%d)", i)
		time.Sleep(time.Second * 2)
		i++
	}

}
