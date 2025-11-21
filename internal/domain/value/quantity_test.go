package value_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

func TestNewQuantity(t *testing.T) {
	tests := map[string]struct {
		input     int
		want      value.Quantity
		wantError error
	}{
		"valid quantity": {
			input: 1,
			want:  value.Quantity(1),
		},
		"valid quantity 10": {
			input: 10,
			want:  value.Quantity(10),
		},
		"maximum allowed quantity": {
			input: 1000,
			want:  value.Quantity(1000),
		},
		"zero quantity": {
			input:     0,
			wantError: value.ErrQuantityInvalid,
		},
		"negative quantity": {
			input:     -1,
			wantError: value.ErrQuantityInvalid,
		},
		"quantity too large": {
			input:     1001,
			wantError: value.ErrQuantityTooLarge,
		},
		"very large quantity": {
			input:     9999,
			wantError: value.ErrQuantityTooLarge,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := value.NewQuantity(tt.input)

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

func TestQuantity_Int(t *testing.T) {
	tests := map[string]struct {
		quantity value.Quantity
		want     int
	}{
		"quantity 1": {
			quantity: value.Quantity(1),
			want:     1,
		},
		"quantity 50": {
			quantity: value.Quantity(50),
			want:     50,
		},
		"quantity 1000": {
			quantity: value.Quantity(1000),
			want:     1000,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := tt.quantity.Int()

			require.Equal(t, tt.want, result)
		})
	}
}

func TestQuantity_Add(t *testing.T) {
	tests := map[string]struct {
		quantity  value.Quantity
		other     value.Quantity
		want      value.Quantity
		wantError error
	}{
		"add valid quantities": {
			quantity: value.Quantity(5),
			other:    value.Quantity(3),
			want:     value.Quantity(8),
		},
		"add to reach maximum": {
			quantity: value.Quantity(500),
			other:    value.Quantity(500),
			want:     value.Quantity(1000),
		},
		"add resulting in too large quantity": {
			quantity:  value.Quantity(600),
			other:     value.Quantity(500),
			wantError: value.ErrQuantityTooLarge,
		},
		"add 1 to maximum": {
			quantity:  value.Quantity(1000),
			other:     value.Quantity(1),
			wantError: value.ErrQuantityTooLarge,
		},
		"add zero": {
			quantity: value.Quantity(5),
			other:    value.Quantity(0),
			want:     value.Quantity(5),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.quantity.Add(tt.other)

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