package stream

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"time"
)

/*

> go run -race cmd/my/program.go | beam

*/

// Stream allows to stream logs via a connection
// to a receiving endpoint
// scotty currently supports unix socket and tcp socket
// streams
type Streamer interface {
	// Stream is a blocking operation which process and streams
	// received logs via os.Stdin to scotty
	Stream(stop <-chan struct{}) error
}

type stream struct {
	label  string
	reader *bufio.Reader
	sock   net.Conn
}

func New(label string, conn net.Conn) Streamer {

	// in order to distinct between multiple streams
	// generate a random value if not set
	if len(label) == 0 {
		label = randLabel(16)
	}

	return &stream{
		label:  label,
		reader: bufio.NewReaderSize(os.Stdin, 32),
		sock:   conn,
	}
}

func (str stream) Stream(stop <-chan struct{}) error {
	defer str.close()

	for {
		select {
		case <-stop:
			return nil
		default:

			log, err := str.reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return fmt.Errorf("unable to read from os.Stdin: %w", err)
			}

			if _, err := str.sock.Write(log); err != nil {
				return fmt.Errorf("unable to write to scotty: %v", err)
			}
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (str stream) read() string {
	s, _ := str.reader.ReadString('\n')
	return s
}

// const (
// 	// synFlag is used to say hi to scotty after connecting
// 	syncFlag = "SYNC"
// )

// func (s *stream) connect(ipc string) (io.WriteCloser, error) {

// 	var err error
// 	if s.sock, err = net.Dial("unix", ipc); err != nil {
// 		return nil, fmt.Errorf("unable to connect to unix socket %q: %w", ipc, err)
// 	}

// 	// send hello SYN flag to scotty which includes meta-data about the beam
// 	// such as the stream name (if provided)
// 	if err := s.sync(); err != nil {
// 		return nil, err
// 	}

// 	return s, nil
// }

// // sync tells the running scotty process that a new stream is about to
// // stream logs. Within the message certain meta-data such as the stream name
// // can be announced to scotty
// func (s *stream) sync() error {

// 	var syncMsg = []byte(fmt.Sprintf("%s;stream=%s", syncFlag, s.label))

// 	if _, err := s.Write(syncMsg); err != nil {
// 		return fmt.Errorf("beam is unable to sync with scotty: %w", err)
// 	}

// 	return nil
// }

// func (s *stream) Write(b []byte) (int, error) {
// 	b = append(b, '\n')
// 	return s.sock.Write(b)
// }

func (s *stream) close() error {
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
