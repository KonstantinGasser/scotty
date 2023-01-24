package multiplexer

// any error captured while
// adding/reading from a stream
type Error error

// any new stream which is connecting
// to scotty represented by a name provided
// via beam -label
type Subscriber string

// any stream returning an io.EOF therefore
// closing the connection or any other reason
// leading to a connection drop
type Unsubscribe string

// any message send by any connected
// stream
type Message struct {
	Label string
	Data  []byte
}
