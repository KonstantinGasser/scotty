package multiplexer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

var (
	newLine = byte(10) // -> \n
)

type stream struct {
	label  string
	msgs   chan<- Message
	reader net.Conn
}

func newStream(conn net.Conn, msgs chan<- Message) (*stream, error) {

	s := stream{
		msgs:   msgs,
		reader: conn,
	}

	if err := s.waitForSync(); err != nil {
		return nil, err
	}

	return &s, nil
}

var (
	ErrConnDropped = fmt.Errorf("beam closed the connection")
)

func (s *stream) handle() error {
	defer s.reader.Close()

	var buf = bufio.NewReader(s.reader)
	for {

		msg, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return Error(fmt.Errorf("unable to read from %q: %w", s.label, err))
		}

		if len(msg) <= 0 {
			continue
		}

		if len(msg) == 1 && msg[0] == newLine {
			continue
		}

		if msg[len(msg)-1] == newLine {
			msg = msg[:len(msg)-1]
		}

		s.msgs <- Message{
			Label: s.label,
			Data:  msg,
		}
	}
	// if we reach this line the EOF broke the look and it is safe
	// to assume that the client closed the connection
	return ErrConnDropped
}

func (s *stream) waitForSync() error {

	// error out if beam is not able to send the sync
	// message within 5 seconds
	// if err := s.reader.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
	// 	return fmt.Errorf("timeout while waiting for SYNC message of beam: %w", err)
	// }

	var buf = bufio.NewReader(s.reader)

	// block until deadline is reached waiting
	// for the sync
	msg, err := buf.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("unable to read SYNC message from connecting beam: %w", err)
	}

	type (
		metadata struct {
			Label string `json:"label"`
		}
	)

	var meta metadata
	if err := json.Unmarshal(msg, &meta); err != nil {
		return fmt.Errorf("SYNC message malformed: %w", err)
	}

	s.label = meta.Label
	return nil
}
