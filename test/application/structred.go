package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"go.uber.org/zap"
)

const (
	logErr = iota
	logPanic
	logWarn
	logDebug
	logInfo
)

// literally just an var with the int 0 (zero)
// which however tricks the linter/compiler to
// not asume a division by zero error
var zeroInt int

func main() {

	delay := flag.Int("delay", 100, "max rand millisecond before the next log is generated")

	flag.Parse()

	rand.Seed(time.Now().Unix())
	sl, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	sl.Info("Hello Scotty", zap.String("state", "ready to beam"), zap.Int("index", 0))

	var i int = 1
	for {
		i++
		time.Sleep(time.Millisecond * time.Duration(*delay))
		p := rand.Float64()
		handleLog(sl, i, p)
	}
}

func handleLog(logger *zap.Logger, i int, p float64) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Println(string(buf))
		}
	}()

	switch {
	// 40% of the time we log an error
	case p > 0.4:
		logger.Error("unable to do X", zap.Int("index", i), zap.Error(fmt.Errorf("unable to do X")), zap.Time("ts", time.Now()))
		break
	// 30% of the time we log an info
	case p > 0.3:
		logger.Info("route XYZ called", zap.Int("index", i), zap.String("remote-id", "127.0.0.1:6598"))
		break
	// 20% of the time we log an warning
	case p > 0.2:
		logger.Warn("caution this indicates X", zap.Int("index", i), zap.Time("ts", time.Now()))
		break
		// with the lowest probabiliy if meet we cause
		// panic by dividing through zero
		// case p > 0.1:
		// 	fmt.Print(4 / zeroInt)
		// 	break
	}
}
