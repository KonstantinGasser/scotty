package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"time"
)

/*g

> go run -race cmd/my/program.go | beam

beam consumes each log line and process it and sends to to scotty.

The processing unit tries parsing the log in into json (or uses some heuristic to determine if
it is worth trying to parse to to structures logs). After parsing the unit writes the result (parsed or raw)
to some writer


:for
::read io from stdin
:::process read data
::::send processed data
:next iteration
*/

// Stream allows to stream logs via a connection
// to a receving endpoint
// scotty currently supports unix socket and tcp socket
// streams
type Stream interface {
	// Connect should connect to the request scotty process
	// how depends on which type is chosen (uinx, tcp)
	Connect(addr string) error
	// Close the stream and the underlying net.Conn
	Close() error
	// io.Writer interface implemented by net.Conn as such only a wrapper
	Write(b []byte) (int, error)
}

func newStream(protocol string, label string) (Stream, error) {

	// in order to distinct between multiple streams
	// generate a random value if not set
	if len(label) == 0 {
		label = randLabel(8)
	}

	switch protocol {
	case "unix":
		return &socket{
			stream: label,
		}, nil
	case "tcp":
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("unknown protocol: %q", protocol)
	}
}

type socket struct {
	stream string
	sock   net.Conn
}

const (
	// synFlag is used to say hi to scotty after connecting
	syncFlag = "SYNC"
)

func (s *socket) Connect(ipc string) error {

	var err error
	if s.sock, err = net.Dial("unix", ipc); err != nil {
		return fmt.Errorf("unable to connect to unix socket %q: %w", ipc, err)
	}

	// send hello SYN flag to scotty which includes meta-data about the beam
	// such as the stream name (if provided)
	if err := s.sync(); err != nil {
		return err
	}

	return nil
}

// sync tells the running scotty process that a new stream is about to
// stream logs. Within the message certain meta-data such as the stream name
// can be announced to scotty
func (s *socket) sync() error {

	var syncMsg = []byte(fmt.Sprintf("%s;stream=%s", syncFlag, s.stream))

	if _, err := s.Write(syncMsg); err != nil {
		return fmt.Errorf("beam is unable to sync with scotty: %w", err)
	}

	return nil
}

func (s *socket) Write(b []byte) (int, error) {
	b = append(b, '\n')
	return s.sock.Write(b)
}

func (s *socket) Close() error {
	if s.sock == nil {
		return nil
	}
	return s.sock.Close()
}

var (
	letters = []byte("abcdefghijklmnopqrstuvwxyz")
)

func randLabel(size int) string {
	rand.Seed(time.Now().Unix())

	var out bytes.Buffer

	for i := 0; i < size; i++ {
		out.WriteByte(letters[rand.Intn(26)])
	}

	return out.String()
}
