package models

import "github.com/pocketbase/pocketbase/tools/types"

var _ Model = (*Request)(nil)

type Error struct {
	BaseModel

	doPanic bool

	Error string `json:"error" db:"error"`
	File  string `json:"file" db:"file"`
	Line  int    `json:"line" db:"line"`
	Fatal bool   `json:"fatal" db:"fatal"`

	Meta types.JsonMap `json:"meta" db:"meta"`
}

func NewFatalError(doPanic ...bool) *Error {
	return &Error{
		Fatal:   true,
		doPanic: len(doPanic) > 0 && doPanic[0],
	}
}

func (m Error) String() string {
	return m.Error
}

func (m *Error) TableName() string {
	return "_errors"
}
