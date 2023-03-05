package filter

// describes any function which can receive a log label
// and the actual log line and based on this information
// decides to include the item or not
type Func func(label string, data []byte) bool

func Default(label string, data []byte) bool {
	return true
}

func WithHighlight(streams ...string) Func {

	return func(label string, data []byte) bool {
		var ok bool = false
		for _, s := range streams {
			if s == label {
				ok = true
			}
		}
		return ok
	}
}
