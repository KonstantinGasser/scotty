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

// unix represents a unix socket connection
// over which logs can be send. unix implements
// io.Writer interface
type unix struct {
	sock net.Conn
}

func newUnix(ipc string) (*unix, error) {

	conn, err := net.Dial("unix", fmt.Sprintf("/tmp/%s.sock", ipc))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to %q socket: %w", ipc, err)
	}

	return &unix{sock: conn}, nil
}

func (ux unix) Write(b []byte) (int, error) {
	return ux.sock.Write(b)
}
