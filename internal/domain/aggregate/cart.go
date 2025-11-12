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
	CartStatusOpen   CartStatus = "OPEN"
	CartStatusClosed CartStatus = "CLOSED"
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
		version:           -1, // -1 indicates new aggregate
		uncommittedEvents: make([]event.Event, 0),
	}
}

func (a *CartAggregate) GetAggregateID() uuid.UUID {
	return a.aggregateID
}

func (a *CartAggregate) GetUserID() uuid.UUID {
	return a.userID
}

func (a *CartAggregate) GetItems() []*entity.CartItem {
	return a.items
}

func (a *CartAggregate) GetStatus() CartStatus {
	return a.status
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

func (a *CartAggregate) IsNew() bool {
	return a.version == -1
}

// ExecuteAddItemToCartCommand handles adding items to cart (creates cart if new)
func (a *CartAggregate) ExecuteAddItemToCartCommand(cmd command.AddItemToCartCommand) error {
	// Check if cart is closed
	if a.status == CartStatusClosed {
		return ErrCartClosed
	}

	// If cart is new, create it first
	if a.IsNew() {
		a.aggregateID = cmd.CartID
		a.userID = cmd.CartID // Simplified: using CartID as UserID for now
		a.status = CartStatusOpen
		a.version = 1

		// Create CartCreated event
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

	// Check if item already exists, update quantity
	for i, item := range a.items {
		if item.ItemID == cmd.ItemID {
			newQuantity, err := value.NewQuantity(item.Quantity.Int() + quantity.Int())
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

	// Add new item
	cartItem := entity.NewCartItem(cmd.ItemID, quantity, price)
	a.items = append(a.items, cartItem)

	a.version++
	evt := event.NewItemAddedToCartEvent(a.aggregateID, a.version, cmd.ItemID, quantity.Int(), price.Float64())
	a.uncommittedEvents = append(a.uncommittedEvents, evt)

	return nil
}

// GetTotalAmount calculates total amount for purchase
func (a *CartAggregate) GetTotalAmount() float64 {
	total := 0.0
	for _, item := range a.items {
		total += item.GetTotal()
	}
	return total
}

// ExecutePurchaseCartCommand handles cart purchase (closes the cart)
func (a *CartAggregate) ExecutePurchaseCartCommand() error {
	if a.IsNew() {
		return errors.UnpermittedOp.New("cannot purchase empty cart")
	}

	if a.status == CartStatusClosed {
		return ErrCartClosed
	}

	if len(a.items) == 0 {
		return errors.UnpermittedOp.New("cannot purchase empty cart")
	}

	a.version++
	totalAmount := a.GetTotalAmount()
	evt := event.NewCartPurchasedEvent(a.aggregateID, a.version, totalAmount)
	a.uncommittedEvents = append(a.uncommittedEvents, evt)

	return nil
}

// Hydration rebuilds the aggregate from events
func (a *CartAggregate) Hydration(events []event.Event) error {
	for _, evt := range events {
		switch e := evt.(type) {
		case *event.CartCreatedEvent:
			a.aggregateID = e.GetAggregateID()
			a.userID = e.GetUserID()
			a.status = CartStatusOpen
			a.version = e.GetVersion()
		case *event.ItemAddedToCartEvent:
			quantity, err := value.NewQuantity(e.GetQuantity())
			if err != nil {
				return err
			}
			price, err := value.NewPrice(e.GetPrice())
			if err != nil {
				return err
			}

			// Update existing item or add new one
			found := false
			for i, item := range a.items {
				if item.ItemID == e.GetItemID() {
					newQuantity, err := value.NewQuantity(item.Quantity.Int() + quantity.Int())
					if err != nil {
						return err
					}
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