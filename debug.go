package main

import (
	"log"
	"fmt"
	"os"
)

var xstd = log.New(os.Stderr, "", log.LstdFlags)

func Debug(level int, format string, v ... interface{}) {
	if opts.debug >= level {
		xstd.Output(2, fmt.Sprintf(format, v...))
	}
}

func DebugFatal(Status int, format string, v ... interface{}) {
	if opts.debug > 0 {
		xstd.Output(2, fmt.Sprintf(format, v...))
		os.Exit(Status)
	}
}

func DebugLn(level int, v ... interface{}) {
	if opts.debug >= level {
		xstd.Output(2, fmt.Sprintln(v...))
	}
}
