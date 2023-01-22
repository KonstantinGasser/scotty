package main

import (
	"fmt"
	"strings"

	"github.com/KonstantinGasser/scotty/ring"
)

func main() {

	buf := ring.New(12)

	payload := []byte(`{"level":"warn","ts":1674335370.996341,"caller":"application/structred.go:34","msg":"caution this indicates X","index":188,"ts":1674335370.996334}`)

	for i := 0; i < 1<<12; i++ {
		buf.Append(payload)
	}

	var w = &strings.Builder{}

	err := buf.Window(w, 50, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

}
