package multiplexer

import (
	"errors"
	"fmt"
	"net"
)

type Socket struct {
	quite <-chan struct{}
	// any error while accepting connections, creating the stream
	// or reading from the stream will be piped to this channel
	// so the UI can display errors
	errors chan<- error
	// any message (exclusive the SYNC message) of a stream
	// will be send through this channel
	messages chan []byte
	// communicate that a new stream has connected
	// to scotty - for now we only pipe the stream label
	// as an information to the UI
	beams chan string

	listener net.Listener
}

func New(q <-chan struct{}, network string, addr string) (*Socket, error) {

	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("unable to start scotty socket: %w", err)
	}

	return &Socket{
		quite:    q,
		errors:   make(chan<- error),
		messages: make(chan []byte),
		beams:    make(chan string),
		listener: ln,
	}, nil
}

func (sock *Socket) Run() {

	go func() {
		<-sock.quite
		sock.listener.Close()
	}()

	for {
		conn, err := sock.listener.Accept()
		if err != nil {
			// call to quite lead to closing of listener
			// scotty is shutting down, break
			if errors.Is(err, net.ErrClosed) {
				break
			}

			// not sure what the best thing would be
			// to do here? But we should show the error to the user?
			continue
		}

		// this can be a blocking operation up to 5 seconds
		// (sync timeout)
		go func(c net.Conn) {
			s, err := newStream(c, sock.errors, sock.messages)
			if err != nil {
				sock.errors <- err
				return
			}
			go s.handle()
			sock.beams <- s.label
		}(conn)
	}
}
