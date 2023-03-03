// attribution: https://github.com/rs/zerolog/
package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Level defines log levels.
type Level int8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
	NoLevel
	Disabled
	TraceLevel Level = -1
)

var (
	// LevelTraceValue is the value used for the trace level field.
	LevelTraceValue = "trace"
	// LevelDebugValue is the value used for the debug level field.
	LevelDebugValue = "debug"
	// LevelInfoValue is the value used for the info level field.
	LevelInfoValue = "info"
	// LevelWarnValue is the value used for the warn level field.
	LevelWarnValue = "warn"
	// LevelErrorValue is the value used for the error level field.
	LevelErrorValue = "error"
	// LevelFatalValue is the value used for the fatal level field.
	LevelFatalValue = "fatal"
	// LevelPanicValue is the value used for the panic level field.
	LevelPanicValue = "panic"

	// LevelFieldMarshalFunc allows customization of global level field marshaling.
	LevelFieldMarshalFunc = func(l Level) string {
		return l.String()
	}
)

func (l Level) String() string {
	switch l {
	case TraceLevel:
		return LevelTraceValue
	case DebugLevel:
		return LevelDebugValue
	case InfoLevel:
		return LevelInfoValue
	case WarnLevel:
		return LevelWarnValue
	case ErrorLevel:
		return LevelErrorValue
	case FatalLevel:
		return LevelFatalValue
	case PanicLevel:
		return LevelPanicValue
	case Disabled:
		return "disabled"
	case NoLevel:
		return ""
	}
	return strconv.Itoa(int(l))
}

// ParseLevel converts a level string into a zerolog Level value.
// returns an error if the input string does not match known values.
func ParseLevel(levelStr string) (Level, error) {
	switch strings.ToLower(levelStr) {
	case LevelFieldMarshalFunc(TraceLevel):
		return TraceLevel, nil
	case LevelFieldMarshalFunc(DebugLevel):
		return DebugLevel, nil
	case LevelFieldMarshalFunc(InfoLevel):
		return InfoLevel, nil
	case LevelFieldMarshalFunc(WarnLevel):
		return WarnLevel, nil
	case LevelFieldMarshalFunc(ErrorLevel):
		return ErrorLevel, nil
	case LevelFieldMarshalFunc(FatalLevel):
		return FatalLevel, nil
	case LevelFieldMarshalFunc(PanicLevel):
		return PanicLevel, nil
	case LevelFieldMarshalFunc(Disabled):
		return Disabled, nil
	case LevelFieldMarshalFunc(NoLevel):
		return NoLevel, nil
	}
	i, err := strconv.Atoi(levelStr)
	if err != nil {
		return NoLevel, fmt.Errorf("Unknown Level String: '%s', defaulting to NoLevel", levelStr)
	}
	if i > 127 || i < -128 {
		return NoLevel, fmt.Errorf("Out-Of-Bounds Level: '%d', defaulting to NoLevel", i)
	}
	return Level(i), nil
}

// UnmarshalText implements encoding.TextUnmarshaler to allow for easy reading from toml/yaml/json formats
func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errors.New("can't unmarshal a nil *Level")
	}
	var err error
	*l, err = ParseLevel(string(text))
	return err
}

// MarshalText implements encoding.TextMarshaler to allow for easy writing into toml/yaml/json formats
func (l Level) MarshalText() ([]byte, error) {
	return []byte(LevelFieldMarshalFunc(l)), nil
}
