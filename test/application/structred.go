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
	for {

		switch rand.Intn(info) {
		case erro:

			sl.Error("unable to do X", zap.Error(fmt.Errorf("unable to do X")), zap.Time("ts", time.Now()))
		case warn:
			sl.Warn("caution this indicates X", zap.Time("ts", time.Now()))
		case debug:
			sl.Debug("depth of the tree", zap.Int("depth", rand.Int()))
		case info:
			sl.Info("route XYZ called", zap.String("remote-id", "127.0.0.1:6598"))
		}

		time.Sleep(time.Millisecond * time.Duration(rand.Intn(250)))
	}
}
