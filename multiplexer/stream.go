package multiplexer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

type stream struct {
	label  string
	errs   chan<- error
	msgs   chan<- []byte
	reader net.Conn
}

func newStream(conn net.Conn, errs chan<- error, msgs chan<- []byte) (*stream, error) {

	s := stream{
		errs:   errs,
		msgs:   msgs,
		reader: conn,
	}

	if err := s.waitForSync(); err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *stream) handle() {

	defer s.reader.Close()

	var buf = bufio.NewReader(s.reader)
	for {

		// unblock I/O if the connection fails to send a message within a second and continue
		// in the next round
		if timeout := s.reader.SetReadDeadline(time.Now().Add(time.Second * 1)); timeout != nil {
			continue
		}

		msg, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// here it would be nice to notify
				// the user through the scotty ui that
				// the stream has disconnected/closed
				break
			}
			s.errs <- fmt.Errorf("unable to read from %q: %w", s.label, err)
			return
		}
		s.msgs <- msg
	}
}

func (s *stream) waitForSync() error {

	// error out if beam is not able to send the sync
	// message within 5 seconds
	if err := s.reader.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
		return fmt.Errorf("timeout while waiting for SYNC message of beam: %w", err)
	}

	var buf = bufio.NewReader(s.reader)

	// block until deadline is reached waiting
	// for the sync
	msg, err := buf.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("unable to read SYNC message from connecting beam: %w", err)
	}

	type (
		metadata struct {
			Label string
		}
	)

	var meta metadata
	if err := json.Unmarshal(msg, &meta); err != nil {
		return fmt.Errorf("SYNC message malformed: %w", err)
	}

	s.label = meta.Label
	return nil
}
