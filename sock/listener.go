package sock

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

func Open(network string, addr string) (net.Listener, error) {
	return net.Listen(network, addr)
}

func Listen(ln net.Listener, w io.Writer, stop <-chan struct{}) {

	go func() {
		<-stop
		if err := ln.Close(); err != nil {
			fmt.Println("Closing the socket was not successful, in case you have issues re-running scotty delete the created socket")
		}
	}()

	for {

		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("unable to access connection: %v", err)
			return
		}

		go read(conn, w)

	}
}

func read(conn net.Conn, w io.Writer) {

	var buf = bufio.NewReader(conn)

	for {

		msg, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// write closing message to io.Writer
				break
			}
			break
		}

		w.Write(msg)
	}
}
