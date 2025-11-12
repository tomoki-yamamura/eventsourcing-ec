package value

import (
	"strings"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

var (
	ErrUserIDEmpty   = errors.InvalidParameter.New("user_id cannot be empty")
	ErrUserIDTooLong = errors.InvalidParameter.New("user_id cannot exceed 128 characters")
)

type UserID string

func NewUserID(id string) (UserID, error) {
	trimmed := strings.TrimSpace(id)

	if trimmed == "" {
		return "", ErrUserIDEmpty
	}

	if len(trimmed) > 128 {
		return "", ErrUserIDTooLong
	}

	return UserID(trimmed), nil
}

func (u UserID) String() string {
	return string(u)
}
