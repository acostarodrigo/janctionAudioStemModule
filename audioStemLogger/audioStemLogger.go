package audioStemLogger

import (
	"fmt"
	"log"
	"os"
)

// ANSI escape codes for colors and formatting
const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorBlue  = "\033[34m"
	colorBold  = "\033[1m"
)

// Log levels
const (
	LevelInfo  = 1
	LevelDebug = 2
	LevelError = 3
)

// AudioStemLogger defines a custom logger for the module
type AudioStemLogger struct {
	logger   *log.Logger
	logLevel int
}

// NewAudioStemLogger creates a new instance of the logger with a specified log level
func NewAudioStemLogger(level int) *AudioStemLogger {
	return &AudioStemLogger{
		logger:   log.New(os.Stdout, colorRed+"[AudioStem] "+colorReset, log.LstdFlags),
		logLevel: level,
	}
}

// GlobalLogger provides a globally accessible logger instance with default level INFO
var Logger = NewAudioStemLogger(LevelInfo)

// Info logs informational messages (Bold Green) if log level allows
func (v *AudioStemLogger) Info(format string, args ...interface{}) {
	if v.logLevel <= LevelInfo {
		v.logger.Println(colorBold + colorGreen + "INFO: " + colorReset + fmt.Sprintf(format, args...))
	}
}

// Error logs error messages (Bold Red) if log level allows
func (v *AudioStemLogger) Error(format string, args ...interface{}) {
	if v.logLevel <= LevelError {
		v.logger.Println(colorBold + colorRed + "ERROR: " + fmt.Sprintf(format, args...) + colorReset)
	}
}

// Debug logs debug messages (Bold Blue) if log level allows
func (v *AudioStemLogger) Debug(format string, args ...interface{}) {
	if v.logLevel <= LevelDebug {
		v.logger.Println(colorBold + colorBlue + "DEBUG: " + fmt.Sprintf(format, args...) + colorReset)
	}
}
