package value_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

func TestNewUserID(t *testing.T) {
	tests := map[string]struct {
		input     string
		want      value.UserID
		wantError error
	}{
		"valid user ID": {
			input: "user123",
			want:  value.UserID("user123"),
		},
		"user ID with spaces": {
			input: "  user123  ",
			want:  value.UserID("user123"),
		},
		"empty user ID": {
			input:     "",
			wantError: value.ErrUserIDEmpty,
		},
		"user ID with only spaces": {
			input:     "   ",
			wantError: value.ErrUserIDEmpty,
		},
		"too long user ID": {
			input:     strings.Repeat("a", 129),
			wantError: value.ErrUserIDTooLong,
		},
		"exactly 128 characters": {
			input: strings.Repeat("a", 128),
			want:  value.UserID(strings.Repeat("a", 128)),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := value.NewUserID(tt.input)

			if tt.wantError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, result)
			}
		})
	}
}
