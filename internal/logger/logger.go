package logger

import (
	"io"
	"log"
	"os"
	"sync"
)

type Level int

const (
	LevelTrace Level = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

var (
	globalLogger *Logger
	once         sync.Once
)

type Logger struct {
	logger *log.Logger
	level  Level
	mu     sync.RWMutex
}

func NewLogger(output io.Writer, prefix string, flag int, level Level) *Logger {
	return &Logger{
		logger: log.New(output, prefix, flag),
		level:  level,
	}
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) log(level Level, prefix string, format string, args ...interface{}) {
	l.mu.RLock()
	currentLevel := l.level
	l.mu.RUnlock()

	if level >= currentLevel {
		if len(args) > 0 {
			l.logger.Printf(prefix+format, args...)
		} else {
			l.logger.Print(prefix + format)
		}
	}
}

func (l *Logger) Trace(format string, args ...interface{}) {
	l.log(LevelTrace, "[TRACE] ", format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, "[INFO] ", format, args...)
}

func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(LevelWarning, "[WARN] ", format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, "[ERROR] ", format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, "[FATAL] ", format, args...)
	os.Exit(1)
}

func Init() {
	once.Do(func() {
		globalLogger = NewLogger(os.Stdout, "", log.LstdFlags|log.Lshortfile, LevelInfo)
	})
}

func SetLevel(level Level) {
	Init()
	globalLogger.SetLevel(level)
}

func SetOutput(output io.Writer) {
	Init()
	globalLogger.logger.SetOutput(output)
}

func Trace(format string, args ...interface{}) {
	Init()
	globalLogger.Trace(format, args...)
}

func Info(format string, args ...interface{}) {
	Init()
	globalLogger.Info(format, args...)
}

func Warning(format string, args ...interface{}) {
	Init()
	globalLogger.Warning(format, args...)
}

func Error(format string, args ...interface{}) {
	Init()
	globalLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	Init()
	globalLogger.Fatal(format, args...)
}

func Printf(format string, args ...interface{}) {
	Init()
	globalLogger.Info(format, args...)
}

func Println(args ...interface{}) {
	Init()
	if len(args) > 0 {
		globalLogger.logger.Println(args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	Init()
	globalLogger.Fatal(format, args...)
}
