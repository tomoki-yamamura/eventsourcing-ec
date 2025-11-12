package value

import (
	"strings"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

var (
	ErrTodoTextEmpty   = errors.InvalidParameter.New("todo_text cannot be empty")
	ErrTodoTextTooLong = errors.InvalidParameter.New("todo_text cannot exceed 256 characters")
)

type TodoText string

func NewTodoText(text string) (TodoText, error) {
	trimmed := strings.TrimSpace(text)

	if trimmed == "" {
		return "", ErrTodoTextEmpty
	}

	if len(trimmed) > 256 {
		return "", ErrTodoTextTooLong
	}

	return TodoText(trimmed), nil
}

func (t TodoText) String() string {
	return string(t)
}
