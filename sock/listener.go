package sock

import (
	"fmt"
	"net"
)

func Open(network string, addr string) (net.Listener, error) {
	return net.Listen(network, addr)
}

func Listen(ln net.Listener, message chan<- net.Conn, stop <-chan struct{}) {

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

		message <- conn
	}
}
