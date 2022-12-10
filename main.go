package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	listener, err := net.Listen("unix", "/tmp/scotty.sock")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	sig := make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func(s <-chan os.Signal) {
		<-s
		listener.Close()
		fmt.Println("graceful shutdown")
		os.Exit(0)
	}(sig)

	fmt.Println("Listening on unix socket: /tmp/scotty.sock")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("unable to accept connect: %v\n", err)
			return
		}

		go func(c net.Conn) {
			defer c.Close()

			for {

				buf := make([]byte, 1024)
				if _, err := c.Read(buf); err != nil {
					fmt.Printf("unable to read from client: %v\n", err)
					break
				}

				fmt.Printf("[scotty] [conn=%v] %s\n", conn, string(buf))
			}
		}(conn)
	}
}
