package models

import (
	"github.com/pocketbase/pocketbase/tools/types"
)

var _ Model = (*Request)(nil)

type Log struct {
	BaseModel

	Level LogLevel `json:"level" db:"level"`

	Message string        `json:"message" db:"message"`
	Meta    types.JsonMap `json:"meta" db:"meta"`
}

func (m Log) String() string {
	return m.Message
}

func (m *Log) TableName() string {
	return "_logs"
}
