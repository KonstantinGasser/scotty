package pager

import (
	"bytes"
	"testing"
)

func TestSplitfunc(t *testing.T) {

	tt := []struct {
		name  string
		input []byte
		width int
	}{
		{
			name:  "single - no split required",
			input: []byte("hello world, this is a line"),
			width: 100, // 100 bytes in one line allowed,
		},
		{
			name:  "split in input in two lines",
			input: []byte("this is suppose to be a longer line, more test, more, more more..."),
			width: 5, // line must only be 5 bytes long including \n
		},
		{
			name:  "split in input in two lines",
			input: []byte("this is suppose to be a longer line, more test, more, more more...\nnow we are cheeky with a new line char"),
			width: 5, // line must only be 5 bytes long including \n
		},
		{
			name:  "split dummy log line - no \\n chars included",
			input: []byte(`{"level":"error","ts":1675111103.747203,"caller":"application/structred.go:32","msg":"unable to do X","index":31,"error":"unable to do X","ts":1675111103.747197,"stacktrace":"main.main\t/Users/konstantingasser/coffeecode/scotty/test/application/structred.go:32runtime.main\t/usr/local/go/src/runtime/proc.go:250"}`),
			width: 50, // line must only be 50 bytes long including \n
		},
		{
			name:  "split dummy log line - with \\n chars included",
			input: []byte(`{"level":"error","ts":1675111103.747203,"caller":"application/structred.go:32","msg":"unable to do X","index":31,"error":"unable to do X","ts":1675111103.747197,"stacktrace":"main.main\n\t/Users/konstantingasser/coffeecode/scotty/test/application/structred.go:32\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:250"}`),
			width: 50, // line must only be 50 bytes long including \n
		},
	}

	for _, tc := range tt {
		var out = make([]byte, 0, len(tc.input))
		result := splitfunc(tc.width, tc.input, out)

		lines := bytes.Split(result, []byte("\n"))
		for _, line := range lines {
			if len(line) > tc.width {
				t.Fatalf("[%s] splitfunc create line longer then width(%d).\ngot: %d\nline:%s", tc.name, tc.width, len(line), line)
			}
		}
	}
}
