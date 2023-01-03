package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type stream struct {
	label  string
	writer io.Writer
	close  func() error
}

// newStream returns a new stream for beaming of logs. Logs are read in from
// os.Stdin therefore 2>&1 might need to be used to pipe os.Stderr -> os.Stdout.
// If beam is started as a daemon logs read in from os.Stdin are printed to os.Stdout
func newStream(label string, proto string, addr string, asDaemon bool) (*stream, error) {
	wc, err := connect(proto, addr)
	if err != nil {
		return nil, err
	}

	var writers = []io.Writer{wc}
	if asDaemon {
		writers = append(writers, os.Stdout)
	}

	writer := io.MultiWriter(writers...)
	s := stream{
		label:  label,
		writer: writer,
	}
	// capture the close of the io.WriteCloser
	// in order keep things clean
	s.close = wc.Close

	return &s, nil
}

func (s stream) beam(quite <-chan struct{}) {
	defer s.close()

	var reader = bufio.NewReader(os.Stdin)

	for {
		select {
		case <-quite:
			return
		default:
			b, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Printf("unable to read log-line:\n\n%v", err)
				return
			}

			if _, err := s.writer.Write(b); err != nil {
				fmt.Printf("unable to beam log-line to scotty:\n\n%v", err)
				return
			}
		}
	}
}
