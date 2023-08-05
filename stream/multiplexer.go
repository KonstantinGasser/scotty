package stream

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type Consumer interface {
	Errors() <-chan Error
	Messages() <-chan Message
	Subscribers() <-chan Subscriber
	Unsubscribers() <-chan Unsubscribe
}

type Listener struct {
	quite <-chan struct{}
	// any error while accepting connections, creating the stream
	// or reading from the stream will be piped to this channel
	// so the UI can display errors
	errors chan Error
	// any message (exclusive the SYNC message) of a stream
	// will be send through this channel
	messages chan Message
	// communicate that a new stream has connected
	// to scotty - for now we only pipe the stream label
	// as an information to the UI
	subscribe chan Subscriber
	// guards subscriber index
	mtx sync.RWMutex
	// keep an index of labels from streams.
	// duplicated stream labels are not allowed and
	// will lead to silently dropping the connection.
	// The beam command will receive an EOF while the user
	// will be displayed the error
	subscribers map[string]struct{}
	// on client EOF or a read error the stream is closed
	// and the event is propagated using this channel
	unsubscribe chan Unsubscribe

	listener net.Listener
}

func New(q <-chan struct{}, network string, addr string) (*Listener, error) {

	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("unable to start scotty with this network/addrr configurations.\n Make sure no other instance is running on this network/addrr.\nPlease see also the exact network error:\n\t:%v", err)
	}

	return &Listener{
		quite:       q,
		errors:      make(chan Error),
		messages:    make(chan Message, 1000), // TODO @KonstantinGasser: does this channel acutally needs to be buffered or are we just fixing symptoms?
		subscribe:   make(chan Subscriber),
		subscribers: make(map[string]struct{}),
		unsubscribe: make(chan Unsubscribe),
		listener:    ln,
	}, nil
}

func (ln *Listener) Run() {

	go func() {
		<-ln.quite
		ln.listener.Close()
	}()

	for {
		conn, err := ln.listener.Accept()
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
			s, err := newStream(c, ln.messages)
			if err != nil {
				ln.errors <- err
				return
			}
			// check for duplicated beams
			ln.mtx.Lock()
			if _, ok := ln.subscribers[s.label]; ok {
				ln.errors <- fmt.Errorf("the label %q is already used by another stream", s.label)
				ln.mtx.Unlock()

				return
			} else {
				ln.subscribers[s.label] = struct{}{}
				ln.mtx.Unlock()

			}
			ln.subscribe <- Subscriber(s.label)

			// blocking operation until error or EOF of client
			if err := s.handle(); err != nil {
				if err == ErrConnDropped {
					ln.mtx.Lock()
					delete(ln.subscribers, s.label)
					ln.mtx.Unlock()

					ln.unsubscribe <- Unsubscribe(s.label)
					return
				}
				ln.errors <- err
				return
			}

		}(conn)
	}
}

func (ln *Listener) Errors() <-chan Error              { return ln.errors }
func (ln *Listener) Messages() <-chan Message          { return ln.messages }
func (ln *Listener) Subscribers() <-chan Subscriber    { return ln.subscribe }
func (ln *Listener) Unsubscribers() <-chan Unsubscribe { return ln.unsubscribe }
