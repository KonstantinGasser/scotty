package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
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
	unstructured := flag.Bool("unstructured", false, "if set includes non JSON (structured) logs to be printed to stderr")

	flag.Parse()

	rand.Seed(time.Now().Unix())
	sl, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	sl.Info("Hello Scotty", zap.String("state", "ready to beam"), zap.Int("index", 0))

	var i int = 0
	for {
		i++
		time.Sleep(time.Millisecond * time.Duration(*delay))
		p := rand.Float64()
		handleLog(sl, *unstructured, i, p)
	}
}

func handleLog(logger *zap.Logger, allowUnstructured bool, i int, p float64) {
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
		if allowUnstructured && rand.Float64() > 0.5 {
			logrus.Errorf("msg=unable to do X, error=%v, index=%d", fmt.Errorf("unable to do X"), i)
			break
		}
		logger.Error("unable to do X", zap.Int("index", i), zap.Error(fmt.Errorf("unable to do X")), zap.Time("ts", time.Now()))
		break
	// 30% of the time we log an info
	case p > 0.3:
		if allowUnstructured && rand.Float64() > 0.5 {
			logrus.Infof("msg=route XYZ called, index=%d", i)
			break
		}
		logger.Info("route XYZ called", zap.Int("index", i), zap.String("remote-id", "127.0.0.1:6598"))
		break
	// 20% of the time we log an warning
	case p > 0.2:
		if allowUnstructured && rand.Float64() > 0.5 {
			logrus.Warnf("msg=caution this indicates X index=%d", i)
			break
		}
		logger.Warn("caution this indicates X", zap.Int("index", i), zap.Time("ts", time.Now()))
		break
		// with the lowest probabiliy if meet we cause
		// panic by dividing through zero
		// case p > 0.1:
		// 	fmt.Print(4 / zeroInt)
		// 	break
	}
}
