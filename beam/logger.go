package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/theckman/yacspin"
)

func info(msg string, args ...interface{}) {
	color.New(color.FgHiBlue).Printf(msg, args...)
}

func fatal(msg string, args ...interface{}) {
	color.New(color.FgRed).Printf(msg, args...)
}

func warn(msg string, args ...interface{}) {
	color.New(color.FgYellow).Printf(msg, args...)
}

var defaultSpinner = yacspin.Config{
	Frequency:     100 * time.Millisecond,
	CharSet:       yacspin.CharSets[24],
	ShowCursor:    true,
	StopCharacter: "",
	// SpinnerAtEnd:  true,
	StopColors: []string{"fgBlue"},
}

var nopStop = func() error { return nil }

func spin(msg string, args ...interface{}) (func() error, error) {

	spinner, err := yacspin.New(defaultSpinner)
	if err != nil {
		info(msg, args...)
		return nopStop, nil
	}

	if err := spinner.Start(); err != nil {
		return nopStop, err
	}
	spinner.Message(fmt.Sprintf(msg, args...))

	return spinner.Stop, nil
}
