package logging

import (
	"io"
	"log"
	"os"
)

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	OFF
	DEFAULT
)

var logger io.Writer

var defaultlevel int

type Logger struct {
	name    string
	level   int
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	error   *log.Logger
}

func (l *Logger) DebugLn(y ...interface{}) {
	if l.level <= DEBUG {
		l.debug.Println(append([]interface{}{l.name + ":"}, y...)...)
	}
}

func (l *Logger) InfoLn(y ...interface{}) {
	if l.level <= INFO {
		l.info.Println(append([]interface{}{l.name + ":"}, y...)...)
	}
}

func (l *Logger) WarnLn(y ...interface{}) {
	if l.level <= WARNING {
		l.warning.Println(append([]interface{}{l.name + ":"}, y...)...)
	}
}

func (l *Logger) ErrorLn(y ...interface{}) {
	if l.level <= ERROR {
		l.error.Println(append([]interface{}{l.name + ":"}, y...)...)
	}
}

func (l *Logger) SetLevel(level int) {
	l.level = level
}

func New(name string, level int) *Logger {
	l := &Logger{}
	l.name = name
	if level == DEFAULT {
		l.level = defaultlevel
	} else {
		l.level = level
	}
	l.debug = log.New(logger, "[  DEBUG  ] ", log.LstdFlags)
	l.info = log.New(logger, "[  INFO   ] ", log.LstdFlags)
	l.warning = log.New(logger, "[ WARNING ] ", log.LstdFlags)
	l.error = log.New(logger, "[  ERROR  ] ", log.LstdFlags)
	return l
}

func InitLogging(level int) {
	if level == DEFAULT {
		level = INFO
	}
	defaultlevel = level
	logfile, err := os.OpenFile("log.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	logger = io.MultiWriter(os.Stdout, logfile)
	// defer logfile.Close()
}

func StrToLvl(name string) int {
	switch name {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warning":
		return WARNING
	case "error":
		return ERROR
	case "off":
		return OFF
	case "default":
		return DEFAULT
	default:
		return DEFAULT
	}
}
