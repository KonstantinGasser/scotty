package stream

import (
	"fmt"
	"net"
)

func Connection(proto string, addr string) (net.Conn, error) {
	switch proto {
	case "unix":
		return net.Dial("unix", addr)
	default:
		return nil, fmt.Errorf("unknown protocol %q", proto)
	}
}
