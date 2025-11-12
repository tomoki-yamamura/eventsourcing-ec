package value_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

func TestNewTodoText(t *testing.T) {
	tests := map[string]struct {
		input     string
		want      value.TodoText
		wantError error
	}{
		"valid text": {
			input: "買い物に行く",
			want:  value.TodoText("買い物に行く"),
		},
		"text with leading and trailing spaces": {
			input: "  勉強する  ",
			want:  value.TodoText("勉強する"),
		},
		"empty string": {
			input:     "",
			wantError: value.ErrTodoTextEmpty,
		},
		"only spaces": {
			input:     "   ",
			wantError: value.ErrTodoTextEmpty,
		},
		"exactly 256 characters": {
			input: strings.Repeat("a", 256),
			want:  value.TodoText(strings.Repeat("a", 256)),
		},
		"over 256 characters": {
			input:     strings.Repeat("a", 257),
			wantError: value.ErrTodoTextTooLong,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := value.NewTodoText(tt.input)
			if tt.wantError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
