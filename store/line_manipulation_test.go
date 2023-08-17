package store

import (
	"testing"

	"github.com/KonstantinGasser/scotty/store/ring"
	"github.com/charmbracelet/lipgloss"
)

var (
	prefix  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render("hello-world")
	divider = " | "
)

const (
	bodyShort   = `time="2023-08-16T19:06:36+02:00" level=error msg="msg=unable to do X, error=unable to do X, index=7"`
	bodyMedium  = `{"level":"error","ts":1692205600.785263,"caller":"application/structred.go:68","msg":"unable to do X","index":16,"error":"unable to do X","ts":1692205600.785238,"stacktrace":"main.handleLog\n\t/Users/konstantingasser/coffecode/scotty/test/application/structred.go:68\nmain.main\n\t/Users/konstantingasser/coffecode/scotty/test/application/structred.go:47\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:250"}`
	bodyLong    = `{"insertId":"42","jsonPayload":{"message":"There was an error in the application","times":"2019-10-12T07:20:50.52Z"},"httpRequest":{"requestMethod":"GET"},"resource":{"type":"k8s_container","labels":{"container_name":"hello-app","pod_name":"helloworld-gke-6cfd6f4599-9wff8","project_id":"stackdriver-sandbox-92334288","namespace_name":"default","location":"us-west4","cluster_name":"helloworld-gke"}},"timestamp":"2020-11-07T15:57:35.945508391Z","severity":"ERROR","labels":{"user_label_2":"value_2","user_label_1":"value_1"},"logName":"projects/stackdriver-sandbox-92334288/logs/stdout","operation":{"id":"get_data","producer":"github.com/MyProject/MyApplication","first":true},"trace":"projects/my-projectid/traces/06796866738c859f2f19b7cfb3214824","sourceLocation":{"file":"get_data.py","line":"142","function":"getData"},"receiveTimestamp":"2020-11-07T15:57:42.411414059Z","spanId":"000000000000004a"}`
	buggyString = `{"level":"error","ts":1692212915.723973,"caller":"application/structred.go:68","msg":"unable to do X","index":52,"error":"unable to do X","ts":1692212915.723955,"stacktrace":"main.handleLog\n\t/Users/konstantingasser/coffecode/scotty/test/application/structred.go:68\nmain.main\n\t/Users/konstantingasser/coffecode/scotty/application/structred.go:47\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:250"}`
)

func makeItem(body string) ring.Item {
	return ring.Item{
		Label:       "hello-world",
		Raw:         prefix + divider + body,
		DataPointer: len(prefix) + len(divider),
	}

}

func TestLineWrap(t *testing.T) {

	tt := []struct {
		name     string
		ttyWidth int
		item     ring.Item
		want     []string
	}{
		{
			name:     "random test logs",
			ttyWidth: 176,
			item:     makeItem(buggyString),
			want: []string{
				`hello-world | {"level":"error","ts":1692212915.723973,"caller":"application/structred.go:68","msg":"unable to do X","index":52,"error":"unable to do X","ts":1692212915.723955,"`,
				`            | stacktrace":"main.handleLog\n\t/Users/konstantingasser/coffecode/scotty/test/application/structred.go:68\nmain.main\n\t/Users/konstantingasser/coffecode/scotty/ap`,
				`            | plication/structred.go:47\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:250"}`,
			},
		},
		{
			name:     "short log line",
			ttyWidth: 45,
			item:     makeItem(bodyShort),
			want: []string{
				`hello-world | time="2023-08-16T19:06:36+02:00`,
				`            | " level=error msg="msg=unable t`,
				`            | o do X, error=unable to do X, index=7"`,
			},
		},
		{
			name:     "medium log line",
			ttyWidth: 65,
			item:     makeItem(bodyMedium),
			want: []string{
				`hello-world | {"level":"error","ts":1692205600.785263,"caller":"a`,
				`            | pplication/structred.go:68","msg":"unable to do X",`,
				`            | "index":16,"error":"unable to do X","ts":1692205600`,
				`            | .785238,"stacktrace":"main.handleLog\n\t/Users/kons`,
				`            | tantingasser/coffecode/scotty/test/application/stru`,
				`            | ctred.go:68\nmain.main\n\t/Users/konstantingasser/c`,
				`            | offecode/scotty/test/application/structred.go:47\nr`,
				`            | untime.main\n\t/usr/local/go/src/runtime/proc.go:250"}`,
			},
		},
		{
			name:     "long log line",
			ttyWidth: 100,
			item:     makeItem(bodyLong),
			want: []string{
				`hello-world | {"insertId":"42","jsonPayload":{"message":"There was an error in the application","tim`,
				`            | es":"2019-10-12T07:20:50.52Z"},"httpRequest":{"requestMethod":"GET"},"resource":{"type`,
				`            | ":"k8s_container","labels":{"container_name":"hello-app","pod_name":"helloworld-gke-6c`,
				`            | fd6f4599-9wff8","project_id":"stackdriver-sandbox-92334288","namespace_name":"default"`,
				`            | ,"location":"us-west4","cluster_name":"helloworld-gke"}},"timestamp":"2020-11-07T15:57`,
				`            | :35.945508391Z","severity":"ERROR","labels":{"user_label_2":"value_2","user_label_1":"`,
				`            | value_1"},"logName":"projects/stackdriver-sandbox-92334288/logs/stdout","operation":{"`,
				`            | id":"get_data","producer":"github.com/MyProject/MyApplication","first":true},"trace":"`,
				`            | projects/my-projectid/traces/06796866738c859f2f19b7cfb3214824","sourceLocation":{"file`,
				`            | ":"get_data.py","line":"142","function":"getData"},"receiveTimestamp":"2020-11-07T15:5`,
				`            | 7:42.411414059Z","spanId":"000000000000004a"}`,
			},
		},
	}

	for _, tc := range tt {
		lines := lineWrap(tc.item, tc.ttyWidth)

		if len(lines) != len(tc.want) {
			t.Fatalf("[%s] number of lines do not match.\n\tWanted: %d\n\tGot: %d", tc.name, len(tc.want), len(lines))
		}

		for i, line := range lines {
			if line != tc.want[i] {
				t.Fatalf("[%s] lines do not match.\n\tWanted: %s\n\tGot: %s", tc.name, tc.want[i], line)
			}
		}

	}
}

// Current benchmark results:
//
// goos: darwin
// goarch: arm64
// pkg: github.com/KonstantinGasser/scotty/store
// BenchmarkLineWrapShort-12     	51005851	       215.4 ns/op	     240 B/op	       3 allocs/op
func BenchmarkLineWrapShort(b *testing.B) {

	ttyWidth := 45
	item := makeItem(bodyShort)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = lineWrap(item, ttyWidth)
	}
}

// Current benchmark results:
//
// goos: darwin
// goarch: arm64
// pkg: github.com/KonstantinGasser/scotty/store
// BenchmarkLineWrapMedium-12    	28089668	       427.9 ns/op	    1616 B/op	       4 allocs/op
func BenchmarkLineWrapMedium(b *testing.B) {

	ttyWidth := 65
	item := makeItem(bodyMedium)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = lineWrap(item, ttyWidth)
	}
}

// Current benchmark results:
//
// goos: darwin
// goarch: arm64
// pkg: github.com/KonstantinGasser/scotty/store
// BenchmarkLineWrapLong-12      	18874477	       630.1 ns/op	    2880 B/op	       4 allocs/op
func BenchmarkLineWrapLong(b *testing.B) {

	ttyWidth := 100
	item := makeItem(bodyLong)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = lineWrap(item, ttyWidth)
	}
}
