package main

import (
	"bufio"
	"encoding/json"
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
func newStream(label string, proto string, addr string, daemon bool) (*stream, error) {
	if len(label) == 0 {
		return nil, fmt.Errorf("please set a label for the stream in order to identify the stream in scotty")
	}

	wc, err := connect(proto, addr)
	if err != nil {
		return nil, err
	}

	var writers = []io.Writer{}
	if !daemon {
		writers = append(writers, os.Stdout)
	}
	writers = append(writers, wc)

	writer := io.MultiWriter(writers...)
	s := stream{
		label:  label,
		writer: writer,
	}
	// capture the close of the io.WriteCloser
	// in order keep things clean
	s.close = wc.Close

	if err := s.sync(); err != nil {
		return nil, err
	}

	return &s, nil
}

// sync sends an initial message to scotty before starting
// to stream logs. Included in the sync message is (for now)
// only the label of the stream
func (s stream) sync() error {

	b, err := json.Marshal(map[string]string{
		"label": s.label,
	})
	if err != nil {
		return fmt.Errorf("malformed SYNC message: %w", err)
	}

	// append \n -> scotty is reading until \n
	b = append(b, '\n')

	if _, err := s.writer.Write(b); err != nil {
		return fmt.Errorf("unable to write SYNC message to scotty: %w", err)
	}

	return nil
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
				printErr("unable to read log from stdin", err)
				return
			}

			if _, err := s.writer.Write(b); err != nil {
				printErr("beam encountered and issue while beaming the logs...", err)
				return
			}
		}
	}
}
