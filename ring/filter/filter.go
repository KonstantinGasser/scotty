package filter

// describes any function which can receive a log label
// and the actual log line and based on this information
// decides to include the item or not
type Func func(label string, data []byte) bool

func WithHighlight(stream string) Func {
	return func(label string, data []byte) bool {
		return label == stream
	}
}
