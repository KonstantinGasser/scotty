package store

import (
	"math/rand"
	"time"
)

func randomIntSlice(size int) []int {
	rand.Seed(time.Now().Unix())

	out := make([]int, size)

	for i := range out {
		out[i] = rand.Int()
	}

	return out
}
