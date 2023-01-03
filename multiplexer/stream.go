package multiplexer

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type stream struct {
	label  string
	errs   chan<- error
	msgs   chan<- interface{}
	reader io.ReadCloser
}

func handleStream(conn net.Conn, errs chan<- error, msgs chan<- interface{}) {

}

func (s *stream) waitForSync() error {

	// error out if beam is not able to send the sync
	// message within 5 seconds
	if err := conn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
		return fmt.Errorf("timeout while waiting for SYNC message of beam: %w", err)
	}

	var buf = bufio.NewReader(s.reader)

	// block until deadline is reached waiting
	// for the sync
	msg, err := buf.ReadString('\n')
	if err != nil {
		return fmt.Errorf("unable to read SYNC message from connecting beam: %w", err)
	}

	meta := strings.Split(msg, ";")

}
