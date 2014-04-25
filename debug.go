package main

import (
	"log"
	"fmt"
	"os"
)

var xstd = log.New(os.Stderr, "", log.LstdFlags)

func Debug(format string, v ... interface{}) {
	if opts.debug == true {
		xstd.Output(2, fmt.Sprintf(format, v...))
	}
}

func DebugFatal(Status int, format string, v ... interface{}) {
	if opts.debug == true {
		xstd.Output(2, fmt.Sprintf(format, v...))
		os.Exit(Status)
	}
}

func DebugLn(v ... interface{}) {
	if opts.debug == true {
		xstd.Output(2, fmt.Sprintln(v...))
	}
}
