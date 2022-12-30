package main

import (
	"fmt"
	"io"
	"net"
)

const (
	protoUnix = "unix"
	protoTCP  = "tcp"
)

// connect returns a new network connection for the given
// protocol and address
func connect(proto string, addr string) (io.WriteCloser, error) {

	switch proto {
	case protoUnix:
		return net.Dial("unix", addr)
	default:
		return nil, fmt.Errorf("unknown protocol %q (must be %q or %q)", proto, protoUnix, protoTCP)
	}
}
