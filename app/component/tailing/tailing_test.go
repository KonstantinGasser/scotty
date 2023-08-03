package tailing

import (
	"testing"
	"time"

	"github.com/KonstantinGasser/scotty/store"
	"github.com/KonstantinGasser/scotty/stream"
)

/*
Current benchmark results:

goos: darwin
goarch: arm64
pkg: github.com/KonstantinGasser/scotty/app/component/tailing
BenchmarkUpdateSingelBeamNoView
BenchmarkUpdateSingelBeamNoView-12    	 3390290	       351.6 ns/op	     275 B/op	       6 allocs/op
*/
func BenchmarkUpdateSingelBeamNoView(b *testing.B) {

	buffer := store.New(2048)
	reader := buffer.NewPager(50, 120, time.Duration(100))

	model := New(reader)

	for i := 0; i < 2048; i++ {
		buffer.Insert("hello-world", 14, []byte(`hello-world | {"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`))
	}

	msg := stream.Message{
		Label: "hello-world",
		Data:  []byte(`hello-world | {"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`),
	}

	for i := 0; i < b.N; i++ {
		model.Update(msg)
	}
}

/*
Current benchmark results:

goos: darwin
goarch: arm64
pkg: github.com/KonstantinGasser/scotty/app/component/tailing
BenchmarkUpdateSingelBeamWithView
BenchmarkUpdateSingelBeamWithView-12    	 1111128	      1051 ns/op	    4203 B/op	       7 allocs/op
*/
func BenchmarkUpdateSingelBeamWithView(b *testing.B) {
	buffer := store.New(2048)
	reader := buffer.NewPager(50, 120, time.Duration(100))

	model := New(reader)

	for i := 0; i < 2048; i++ {
		buffer.Insert("hello-world", 14, []byte(`hello-world | {"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`))
	}

	msg := stream.Message{
		Label: "hello-world",
		Data:  []byte(`hello-world | {"level":"warn","ts":1680212791.946584,"caller":"application/structred.go:39","msg":"caution this indicates X","index":998,"ts":1680212791.946579}`),
	}

	for i := 0; i < b.N; i++ {
		model.Update(msg)
		model.View()
	}
}
