package value

import (
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

var (
	ErrPriceInvalid = errors.InvalidParameter.New("price must be greater than or equal to 0")
	ErrPriceTooLarge = errors.InvalidParameter.New("price cannot exceed 1000000")
)

type Price float64

func NewPrice(price float64) (Price, error) {
	if price < 0 {
		return 0, ErrPriceInvalid
	}

	if price > 1000000 {
		return 0, ErrPriceTooLarge
	}

	return Price(price), nil
}

func (p Price) Float64() float64 {
	return float64(p)
}