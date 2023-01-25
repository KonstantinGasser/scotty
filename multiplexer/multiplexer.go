package multiplexer

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

type Socket struct {
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

func New(q <-chan struct{}, network string, addr string) (*Socket, error) {

	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("unable to start scotty socket: %w", err)
	}

	return &Socket{
		quite:       q,
		errors:      make(chan Error),
		messages:    make(chan Message),
		subscribe:   make(chan Subscriber),
		subscribers: make(map[string]struct{}),
		unsubscribe: make(chan Unsubscribe),
		listener:    ln,
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
			s, err := newStream(c, sock.messages)
			if err != nil {
				sock.errors <- err
				return
			}
			// check for duplicated beams
			sock.mtx.Lock()
			if _, ok := sock.subscribers[s.label]; ok {
				sock.errors <- fmt.Errorf("the label %q is already used by another stream", s.label)
				return
			} else {
				sock.subscribers[s.label] = struct{}{}
			}
			sock.mtx.Unlock()

			sock.subscribe <- Subscriber(s.label)

			// blocking operation until error or EOF of client
			if err := s.handle(); err != nil {
				if err == ErrConnDropped {
					sock.mtx.Lock()
					delete(sock.subscribers, s.label)
					sock.mtx.Unlock()

					sock.unsubscribe <- Unsubscribe(s.label)
					return
				}
				sock.errors <- err
				return
			}

		}(conn)
	}
}

func (sock *Socket) Errors() <-chan Error            { return sock.errors }
func (sock *Socket) Messages() <-chan Message        { return sock.messages }
func (sock *Socket) Subscribe() <-chan Subscriber    { return sock.subscribe }
func (sock *Socket) Unsubscribe() <-chan Unsubscribe { return sock.unsubscribe }
