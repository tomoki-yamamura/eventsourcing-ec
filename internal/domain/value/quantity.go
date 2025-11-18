package value

import (
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

var (
	ErrQuantityInvalid  = errors.InvalidParameter.New("quantity must be greater than 0")
	ErrQuantityTooLarge = errors.InvalidParameter.New("quantity cannot exceed 1000")
)

type Quantity int

func NewQuantity(quantity int) (Quantity, error) {
	if quantity <= 0 {
		return 0, ErrQuantityInvalid
	}

	if quantity > 1000 {
		return 0, ErrQuantityTooLarge
	}

	return Quantity(quantity), nil
}

func (q Quantity) Int() int {
	return int(q)
}

func (q Quantity) Add(other Quantity) (Quantity, error) {
	newValue := q.Int() + other.Int()
	return NewQuantity(newValue)
}
