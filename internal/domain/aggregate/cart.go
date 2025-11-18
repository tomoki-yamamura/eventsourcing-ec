package aggregate

import (
	"github.com/google/uuid"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/entity"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

var (
	ErrItemNotFound = errors.NotFound.New("item not found in cart")
	ErrCartClosed   = errors.UnpermittedOp.New("cart is already purchased")
)

type CartStatus string

const (
	CartStatusOpen      CartStatus = "OPEN"
	CartStatusClosed    CartStatus = "CLOSED"
	CartStatusAbandoned CartStatus = "ABANDONED"
)

type CartAggregate struct {
	aggregateID       uuid.UUID
	userID            uuid.UUID
	items             []*entity.CartItem
	status            CartStatus
	version           int
	uncommittedEvents []event.Event
}

func NewCartAggregate() *CartAggregate {
	return &CartAggregate{
		items:             make([]*entity.CartItem, 0),
		status:            CartStatusOpen,
		version:           -1,
		uncommittedEvents: make([]event.Event, 0),
	}
}

func (a *CartAggregate) GetAggregateID() uuid.UUID {
	return a.aggregateID
}

func (a *CartAggregate) GetVersion() int {
	return a.version
}

func (a *CartAggregate) GetUncommittedEvents() []event.Event {
	return a.uncommittedEvents
}

func (a *CartAggregate) MarkEventsAsCommitted() {
	a.uncommittedEvents = make([]event.Event, 0)
}

func (a *CartAggregate) isNew() bool {
	return a.version == -1
}

func (a *CartAggregate) isCartAvailable() bool {
	return a.status != CartStatusClosed
}

func (a *CartAggregate) ExecuteAddItemToCartCommand(cmd command.AddItemToCartCommand) error {
	if !a.isCartAvailable() {
		return ErrCartClosed
	}

	if a.isNew() {
		a.aggregateID = cmd.CartID
		a.userID = cmd.UserID
		a.status = CartStatusOpen
		a.version = 1

		evt := event.NewCartCreatedEvent(a.aggregateID, a.version, a.userID)
		a.uncommittedEvents = append(a.uncommittedEvents, evt)
	}

	quantity, err := value.NewQuantity(cmd.Quantity)
	if err != nil {
		return err
	}

	price, err := value.NewPrice(cmd.Price)
	if err != nil {
		return err
	}

	for i, item := range a.items {
		if item.ItemID == cmd.ItemID {
			newQuantity, err := item.Quantity.Add(quantity)
			if err != nil {
				return err
			}
			a.items[i] = entity.NewCartItem(cmd.ItemID, newQuantity, price)

			a.version++
			evt := event.NewItemAddedToCartEvent(a.aggregateID, a.version, cmd.ItemID, quantity.Int(), price.Float64())
			a.uncommittedEvents = append(a.uncommittedEvents, evt)
			return nil
		}
	}

	cartItem := entity.NewCartItem(cmd.ItemID, quantity, price)
	a.items = append(a.items, cartItem)

	a.version++
	evt := event.NewItemAddedToCartEvent(a.aggregateID, a.version, cmd.ItemID, quantity.Int(), price.Float64())
	a.uncommittedEvents = append(a.uncommittedEvents, evt)

	return nil
}

func (a *CartAggregate) GetTotalAmount() value.Price {
	total := 0.0
	for _, item := range a.items {
		total += item.GetTotal().Float64()
	}
	totalPrice, _ := value.NewPrice(total)
	return totalPrice
}

func (a *CartAggregate) ExecutePurchaseCartCommand() error {
	if a.isNew() {
		return errors.UnpermittedOp.New("cannot purchase empty cart")
	}

	if !a.isCartAvailable() {
		return ErrCartClosed
	}

	if len(a.items) == 0 {
		return errors.UnpermittedOp.New("cannot purchase empty cart")
	}

	a.version++
	totalAmount := a.GetTotalAmount()
	evt := event.NewCartPurchasedEvent(a.aggregateID, a.version, totalAmount.Float64())
	a.uncommittedEvents = append(a.uncommittedEvents, evt)

	return nil
}

func (a *CartAggregate) Hydration(events []event.Event) error {
	for _, evt := range events {
		switch e := evt.(type) {
		case *event.CartCreatedEvent:
			a.aggregateID = e.GetAggregateID()
			a.userID = e.GetUserID()
			a.status = CartStatusOpen
			a.version = e.GetVersion()
		case *event.ItemAddedToCartEvent:
			quantity, _ := value.NewQuantity(e.GetQuantity())
			price, _ := value.NewPrice(e.GetPrice())

			found := false
			for i, item := range a.items {
				if item.ItemID == e.GetItemID() {
					newQuantity, _ := item.Quantity.Add(quantity)
					a.items[i] = entity.NewCartItem(e.GetItemID(), newQuantity, price)
					found = true
					break
				}
			}
			if !found {
				cartItem := entity.NewCartItem(e.GetItemID(), quantity, price)
				a.items = append(a.items, cartItem)
			}
			a.version = e.GetVersion()
		case *event.CartPurchasedEvent:
			a.status = CartStatusClosed
			a.version = e.GetVersion()
		}
	}
	return nil
}
