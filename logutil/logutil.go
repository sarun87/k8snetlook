package logutil

// Simple extension to Golang log using pointers
// from here: https://stackoverflow.com/a/29556974
// Keeping logging simple. See https://dave.cheney.net/2015/11/05/lets-talk-about-logging
// for an interesting take on logging systems :)

import (
	"fmt"
	"log"
	"os"
)

type MyLogger struct {
	*log.Logger
	level int // One of DEBUG, ERROR, INFO
}

const (
	DEBUG = 1 << iota
	INFO
	ERROR
)

var myLog MyLogger

func Error(format string, v ...interface{}) {
	if myLog.level <= ERROR {
		s := fmt.Sprintf("ERROR: "+format, v...)
		myLog.Output(2, s)
	}
}

func Info(format string, v ...interface{}) {
	if myLog.level <= INFO {
		s := fmt.Sprintf(format, v...)
		myLog.Output(2, s)
	}
}

func Debug(format string, v ...interface{}) {
	if myLog.level <= DEBUG {
		s := fmt.Sprintf("DEBUG: "+format, v...)
		myLog.Output(2, s)
	}
}

func SetLogLevel(lvl int) {
	myLog.level = lvl
	myLog.Logger = log.New(os.Stdout, "", 0)
}
