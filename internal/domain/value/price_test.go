package value_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

func TestNewPrice(t *testing.T) {
	tests := map[string]struct {
		input     float64
		want      value.Price
		wantError error
	}{
		"valid price zero": {
			input: 0.0,
			want:  value.Price(0.0),
		},
		"valid price integer": {
			input: 100.0,
			want:  value.Price(100.0),
		},
		"valid price decimal": {
			input: 99.99,
			want:  value.Price(99.99),
		},
		"valid price with cents": {
			input: 12.34,
			want:  value.Price(12.34),
		},
		"maximum allowed price": {
			input: 1000000.0,
			want:  value.Price(1000000.0),
		},
		"negative price": {
			input:     -0.01,
			wantError: value.ErrPriceInvalid,
		},
		"large negative price": {
			input:     -100.0,
			wantError: value.ErrPriceInvalid,
		},
		"price too large": {
			input:     1000000.01,
			wantError: value.ErrPriceTooLarge,
		},
		"very large price": {
			input:     9999999.99,
			wantError: value.ErrPriceTooLarge,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := value.NewPrice(tt.input)

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

func TestPrice_Float64(t *testing.T) {
	tests := map[string]struct {
		price value.Price
		want  float64
	}{
		"price zero": {
			price: value.Price(0.0),
			want:  0.0,
		},
		"price integer": {
			price: value.Price(100.0),
			want:  100.0,
		},
		"price decimal": {
			price: value.Price(99.99),
			want:  99.99,
		},
		"price with cents": {
			price: value.Price(12.34),
			want:  12.34,
		},
		"maximum price": {
			price: value.Price(1000000.0),
			want:  1000000.0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := tt.price.Float64()

			require.Equal(t, tt.want, result)
		})
	}
}