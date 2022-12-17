package streams

import (
	"errors"
	"fmt"
	"net"
)

type Message struct {
	// Label is the label of the stream
	// message is coming from
	Label string
	// Raw is the actual log line
	Raw string
}

const (
	unset = iota
	open
	closed
)

type Multiplexer struct {
	listener    net.Listener
	messages    chan Message
	subscribers chan string
	errs        chan error
	state       int
}

func (multi Multiplexer) Messages() <-chan Message {
	return multi.messages
}
func (multi Multiplexer) Subscribers() <-chan string {
	return multi.subscribers
}

func (multi Multiplexer) Errors() <-chan error {
	return multi.errs
}

func Open(network string, addr string) (*Multiplexer, error) {

	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("unable to listen on %s:%s: %w", network, addr, err)
	}

	return &Multiplexer{
		listener:    ln,
		messages:    make(chan Message),
		subscribers: make(chan string),
		errs:        make(chan error),
		state:       unset,
	}, nil
}

func (multi Multiplexer) Listen(quite chan struct{}) {

	go func() {
		<-quite
		multi.listener.Close()
	}()

	for {

		conn, err := multi.listener.Accept()
		if err != nil {
			// listener has been closed; application has been closed
			if errors.Is(err, net.ErrClosed) {
				break
			}

			// not sure what to do?
			// error message should somehow end up in the base.Model
			// postponed until base.Model is implemented further
			continue
		}

		stream, err := newStream(conn, multi.errs, multi.messages)
		if err != nil {
			multi.errs <- err
			// again error should go to the model somehow
			continue
		}
		go stream.read()
	}
}
