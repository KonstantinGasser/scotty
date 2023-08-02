package info

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

/*
Current benchmark results:

goos: darwin
goarch: arm64
pkg: github.com/KonstantinGasser/scotty/app/component/info
BenchmarkUpdateWithSingleBeam
BenchmarkUpdateWithSingleBeam-12    	385831722	         3.022 ns/op	       0 B/op	       0 allocs/op
*/
func BenchmarkUpdateWithSingleBeam(b *testing.B) {

	model := New()

	// add a single beam
	model.Update(RequestSubscribe("hello-world", lipgloss.Color("#ffffff")))

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		model.Update(RequestIncrement("hello-world"))
	}
}

/*
Current benchmark results:

goos: darwin
goarch: arm64
pkg: github.com/KonstantinGasser/scotty/app/component/info
BenchmarkViewSingleBeam
BenchmarkViewSingleBeam-12    	 6635584	       167.8 ns/op	     152 B/op	       5 allocs/op
*/
func BenchmarkViewSingleBeam(b *testing.B) {

	model := New()

	model.Update(RequestSubscribe("hello-world", lipgloss.Color("#ffffff")))

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		model.Update(RequestIncrement("hello-world"))
		model.View()
	}
}

/*
Current benchmark results:

goos: darwin
goarch: arm64
pkg: github.com/KonstantinGasser/scotty/app/component/info
BenchmarkUpdateWithMultipleBeam
BenchmarkUpdateWithMultipleBeam-12    	353531521	         3.340 ns/op	       0 B/op	       0 allocs/op
*/
func BenchmarkUpdateWithMultipleBeam(b *testing.B) {

	model := New()

	// add a single beam
	model.Update(RequestSubscribe("hello-world", lipgloss.Color("#ffffff")))
	model.Update(RequestSubscribe("hello-friend", lipgloss.Color("#ffffff")))

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		model.Update(RequestIncrement("hello-world"))
	}
}

/*
Current benchmark results:

goos: darwin
goarch: arm64
pkg: github.com/KonstantinGasser/scotty/app/component/info
BenchmarkViewMultipleBeam
BenchmarkViewMultipleBeam-12    	 6523666	       175.5 ns/op	     152 B/op	       5 allocs/op
*/
func BenchmarkViewMultipleBeam(b *testing.B) {

	model := New()

	model.Update(RequestSubscribe("hello-friend-1", lipgloss.Color("#ffffff")))
	model.Update(RequestSubscribe("hello-friend-2", lipgloss.Color("#ffffff")))
	model.Update(RequestSubscribe("hello-friend-3", lipgloss.Color("#ffffff")))
	model.Update(RequestSubscribe("hello-friend-4", lipgloss.Color("#ffffff")))

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		model.Update(RequestIncrement("hello-friend-1"))
		model.Update(RequestIncrement("hello-friend-2"))
		model.Update(RequestIncrement("hello-friend-3"))
		model.Update(RequestIncrement("hello-friend-4"))
		model.View()
	}
}
