package log

import (
	"log"
	"os"
)

type loglevel int

const (
	INFO  loglevel = 0
	DEBUG loglevel = 1
	TRACE loglevel = 2
)

type Logger struct {
	level loglevel
	info  *log.Logger
	debug *log.Logger
	trace *log.Logger
	errlg *log.Logger
}

func InitLog(level loglevel) Logger {
	info := log.New(os.Stdout, "INF ", log.Lmicroseconds)
	debug := log.New(os.Stdout, "DBG ", log.Lmicroseconds)
	trace := log.New(os.Stdout, "TRC ", log.Lmicroseconds)
	errlg := log.New(os.Stdout, "ERR ", log.Lmicroseconds)

	return Logger{
		level: level,
		info:  info,
		debug: debug,
		trace: trace,
		errlg: errlg,
	}
}

func (l *Logger) Info(v ...any) {
	if l.level >= INFO {
		l.info.Println(v...)
	}
}
func (l *Logger) Infof(format string, v ...any) {
	if l.level >= INFO {
		l.info.Printf(format, v...)
	}
}
func (l *Logger) Debug(v ...any) {
	if l.level >= DEBUG {
		l.debug.Println(v...)
	}
}
func (l *Logger) Debugf(format string, v ...any) {
	if l.level >= DEBUG {
		l.debug.Printf(format, v...)
	}
}
func (l *Logger) Trace(v ...any) {
	if l.level >= TRACE {
		l.trace.Println(v...)
	}
}
func (l *Logger) Tracef(format string, v ...any) {
	if l.level >= TRACE {
		l.trace.Printf(format, v...)
	}
}

func (l *Logger) Error(v ...any) {
	l.errlg.Println(v...)
}
func (l *Logger) Errorf(format string, v ...any) {
	l.errlg.Printf(format, v...)
}
