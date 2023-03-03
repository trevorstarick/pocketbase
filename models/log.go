package models

import "github.com/pocketbase/pocketbase/tools/types"

var _ Model = (*Request)(nil)

// Level defines log levels.
type Level string

const (
	TraceLevel = "trace"
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	FatalLevel = "fatal"
	PanicLevel = "panic"
)

type Log struct {
	BaseModel

	Level Level `json:"level" db:"level"`

	Message string        `json:"message" db:"message"`
	Meta    types.JsonMap `json:"meta" db:"meta"`
}

func (m Log) String() string {
	return m.Message
}

func (m *Log) TableName() string {
	return "_logs"
}
