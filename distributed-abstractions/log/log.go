package log

import (
	"log"
	"os"
)

// logger simply overwrites the default logger since microseconds matter in this
// case.
var logger = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)

func Println(v ...any) {
	logger.Println(v...)
}

func Printf(format string, v ...any) {
	logger.Printf(format, v...)
}
