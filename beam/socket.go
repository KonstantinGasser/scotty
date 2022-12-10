package main

import (
	"fmt"
	"net"
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
	Connect(addr string) error
	Close() error
	Write(b []byte) (int, error)
}

func newStream(protocol string) (Stream, error) {
	switch protocol {
	case "unix":
		return &socket{}, nil
	case "tcp":
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("unknown protocol: %q", protocol)
	}
}

// unix represents a unix socket connection
// over which logs can be send. unix implements
// io.Writer interface
type socket struct {
	sock net.Conn
}

func (s *socket) Connect(ipc string) error {

	var err error
	if s.sock, err = net.Dial("unix", ipc); err != nil {
		return fmt.Errorf("unable to connect to unix socket %q: %w", ipc, err)
	}

	return nil
}

func (s *socket) Write(b []byte) (int, error) {
	return s.sock.Write(b)
}

func (s *socket) Close() error {
	if s.sock == nil {
		return nil
	}
	return s.sock.Close()
}
