package streams

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

const (
	metaLabel = "label"
)

type Stream struct {
	label    string
	conn     net.Conn
	errs     chan<- error
	messages chan<- Message
}

func newStream(conn net.Conn, errs chan<- error, msgs chan<- Message) (*Stream, error) {

	meta, err := resolveSyncMessage(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Stream{
		label:    meta[metaLabel],
		conn:     conn,
		errs:     errs,
		messages: msgs,
	}, nil
}

func (str Stream) read() {
	defer str.Close()

	var buf = bufio.NewReader(str.conn)

	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// here it would be nice to notify the model that a stream has been closed
				// by the remote connection
				return
			}
			// of non EOF error notify model about error somehow and stop reading
			return
		}

		// do something with the "line"
		str.messages <- Message{
			Label: str.label,
			Raw:   line,
		}
	}
}

func (str Stream) Close() error {
	return str.conn.Close()
}

var ErrInvalidSyncMessage = fmt.Errorf("connected stream send invalid sync message. key=value format violation")

// resolveSyncMessage waits for the initial message send by a new stream and resolves
// the SYNC message
func resolveSyncMessage(conn net.Conn) (map[string]string, error) {

	var buf = bufio.NewReader(conn)

	sync, err := buf.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("unable to read sync message from connection: %w", err)
	}

	// sync messages are colon separated key-value pairs
	// with meta-data about the process/connection
	// meta-data:
	// - label=<label:string>

	metadata := strings.Split(sync, ";")

	var out = make(map[string]string)
	for _, meta := range metadata {

		keyVal := strings.Split(meta, "=")
		if len(keyVal) != 2 {
			return nil, fmt.Errorf("invalid sync message: %s", sync) // ErrInvalidSyncMessage
		}

		switch keyVal[0] {
		case metaLabel:
			out[keyVal[0]] = strings.TrimSpace(keyVal[1])
		default:
			continue
		}
	}

	return out, nil
}
