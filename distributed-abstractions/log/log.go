package log

import (
	"log"
	"os"
)

// logger simply overwrites the default logger since microseconds matter in this
// case.
var logger = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
var enabled = true

func Println(v ...any) {
	logger.Println(v...)
}

func Printf(format string, v ...any) {
	if enabled {
		logger.Printf(format, v...)
	}
}
