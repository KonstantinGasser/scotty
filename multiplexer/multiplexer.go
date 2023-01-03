package multiplexer

import (
	"errors"
	"fmt"
	"net"
)

type Socket struct {
	quite    <-chan struct{}
	errors   chan<- error
	messages chan interface{}
	beams    chan interface{}

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
		messages: make(chan interface{}),
		beams:    make(chan interface{}),
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

	}
}

// type Beam struct {
// 	reader io.ReadCloser
// }

func handleBeam(conn net.Conn)
