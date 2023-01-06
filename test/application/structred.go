package main

import (
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

const (
	erro = iota
	warn
	debug
	info
)

func main() {

	rand.Seed(time.Now().Unix())
	sl, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	sl.Info("Hello Scotty", zap.String("state", "ready to beam"))
	var i int
	for {

		switch rand.Intn(info + 1) {
		case erro:
			sl.Error("unable to do X", zap.Int("index", i), zap.Error(fmt.Errorf("unable to do X")), zap.Time("ts", time.Now()))
		case warn:
			sl.Warn("caution this indicates X", zap.Int("index", i), zap.Time("ts", time.Now()))
		case debug:
			sl.Debug("depth of the tree", zap.Int("index", i), zap.Int("depth", rand.Int()))
		case info:
			sl.Info("route XYZ called", zap.Int("index", i), zap.String("remote-id", "127.0.0.1:6598"))
		}

		i++
		time.Sleep(time.Millisecond * 100)
	}
}
