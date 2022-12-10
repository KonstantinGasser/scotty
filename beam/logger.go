package main

import (
	"github.com/fatih/color"
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
