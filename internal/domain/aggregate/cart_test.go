package aggregate_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/aggregate"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
)

func TestCartAggregate_ExecuteAddItemToCartCommand(t *testing.T) {
	cartID := uuid.New()
	userID := uuid.New()
	itemID := uuid.New()

	tests := map[string]struct {
		existingItems []command.AddItemToCartCommand
		isSubmitted   bool
		cmd           command.AddItemToCartCommand
		wantErr       error
		wantEventsLen int
		wantVersion   int
	}{
		"should add first item to new cart": {
			existingItems: nil,
			cmd: command.AddItemToCartCommand{
				CartID:   cartID,
				UserID:   userID,
				ItemID:   itemID,
				Quantity: 1,
				Price:    100.0,
			},
			wantErr:       nil,
			wantEventsLen: 2,
			wantVersion:   2,
		},
		"should add item to existing cart": {
			existingItems: []command.AddItemToCartCommand{
				{
					CartID:   cartID,
					UserID:   userID,
					ItemID:   uuid.New(),
					Quantity: 1,
					Price:    50.0,
				},
			},
			cmd: command.AddItemToCartCommand{
				CartID:   cartID,
				UserID:   userID,
				ItemID:   itemID,
				Quantity: 2,
				Price:    75.0,
			},
			wantErr:       nil,
			wantEventsLen: 1,
			wantVersion:   3,
		},
		"should return error for invalid quantity": {
			existingItems: nil,
			cmd: command.AddItemToCartCommand{
				CartID:   cartID,
				UserID:   userID,
				ItemID:   itemID,
				Quantity: 0,
				Price:    100.0,
			},
			wantErr:       value.ErrQuantityInvalid,
			wantEventsLen: 1,
			wantVersion:   1,
		},
		"should return error for invalid price": {
			existingItems: nil,
			cmd: command.AddItemToCartCommand{
				CartID:   cartID,
				UserID:   userID,
				ItemID:   itemID,
				Quantity: 1,
				Price:    -10.0,
			},
			wantErr:       value.ErrPriceInvalid,
			wantEventsLen: 1,
			wantVersion:   1,
		},
		"should accumulate quantity for same item": {
			existingItems: []command.AddItemToCartCommand{
				{
					CartID:   cartID,
					UserID:   userID,
					ItemID:   itemID,
					Quantity: 2,
					Price:    50.0,
				},
			},
			cmd: command.AddItemToCartCommand{
				CartID:   cartID,
				UserID:   userID,
				ItemID:   itemID,
				Quantity: 3,
				Price:    50.0,
			},
			wantErr:       nil,
			wantEventsLen: 1,
			wantVersion:   3,
		},
		"should return error for submitted cart": {
			existingItems: []command.AddItemToCartCommand{
				{
					CartID:   cartID,
					UserID:   userID,
					ItemID:   uuid.New(),
					Quantity: 1,
					Price:    50.0,
				},
			},
			isSubmitted: true,
			cmd: command.AddItemToCartCommand{
				CartID:   cartID,
				UserID:   userID,
				ItemID:   itemID,
				Quantity: 1,
				Price:    100.0,
			},
			wantErr:       aggregate.ErrCartClosed,
			wantEventsLen: 0,
			wantVersion:   3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cart := aggregate.NewCartAggregate()
			for _, existingCmd := range tt.existingItems {
				cart.ExecuteAddItemToCartCommand(existingCmd)
			}
			cart.MarkEventsAsCommitted()

			if tt.isSubmitted {
				cart.ExecuteSubmitCartCommand(command.SubmitCartCommand{CartID: cartID})
				cart.MarkEventsAsCommitted()
			}

			// Act
			err := cart.ExecuteAddItemToCartCommand(tt.cmd)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, cart.GetUncommittedEvents(), tt.wantEventsLen)
			assert.Equal(t, tt.wantVersion, cart.GetVersion())
		})
	}
}

func TestCartAggregate_ExecuteSubmitCartCommand(t *testing.T) {
	cartID := uuid.New()
	userID := uuid.New()
	itemID := uuid.New()

	tests := map[string]struct {
		existingItems []command.AddItemToCartCommand
		cmd           command.SubmitCartCommand
		wantErr       error
		wantEventsLen int
		wantVersion   int
	}{
		"should return error for new cart": {
			existingItems: nil,
			cmd:           command.SubmitCartCommand{CartID: cartID},
			wantErr:       errors.UnpermittedOp.New("cannot submit empty cart"),
			wantEventsLen: 0,
			wantVersion:   -1,
		},
		"should successfully submit cart with items": {
			existingItems: []command.AddItemToCartCommand{
				{
					CartID:   cartID,
					UserID:   userID,
					ItemID:   itemID,
					Quantity: 1,
					Price:    100.0,
				},
			},
			cmd:           command.SubmitCartCommand{CartID: cartID},
			wantErr:       nil,
			wantEventsLen: 1,
			wantVersion:   3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cart := aggregate.NewCartAggregate()
			for _, existingCmd := range tt.existingItems {
				cart.ExecuteAddItemToCartCommand(existingCmd)
			}
			cart.MarkEventsAsCommitted()

			// Act
			err := cart.ExecuteSubmitCartCommand(tt.cmd)

			// Assert
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot submit empty cart")
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, cart.GetUncommittedEvents(), tt.wantEventsLen)
			assert.Equal(t, tt.wantVersion, cart.GetVersion())
		})
	}
}

func TestCartAggregate_GetTotalAmount(t *testing.T) {
	cartID := uuid.New()
	userID := uuid.New()
	itemID1 := uuid.New()
	itemID2 := uuid.New()

	tests := map[string]struct {
		existingItems []command.AddItemToCartCommand
		want          float64
	}{
		"should return 0 for empty cart": {
			existingItems: nil,
			want:          0.0,
		},
		"should return total for single item": {
			existingItems: []command.AddItemToCartCommand{
				{
					CartID:   cartID,
					UserID:   userID,
					ItemID:   itemID1,
					Quantity: 2,
					Price:    50.0,
				},
			},
			want: 100.0,
		},
		"should return total for multiple items": {
			existingItems: []command.AddItemToCartCommand{
				{
					CartID:   cartID,
					UserID:   userID,
					ItemID:   itemID1,
					Quantity: 2,
					Price:    50.0,
				},
				{
					CartID:   cartID,
					UserID:   userID,
					ItemID:   itemID2,
					Quantity: 3,
					Price:    25.0,
				},
			},
			want: 175.0,
		},
		"should handle accumulated quantity for same item": {
			existingItems: []command.AddItemToCartCommand{
				{
					CartID:   cartID,
					UserID:   userID,
					ItemID:   itemID1,
					Quantity: 2,
					Price:    30.0,
				},
				{
					CartID:   cartID,
					UserID:   userID,
					ItemID:   itemID1,
					Quantity: 3,
					Price:    30.0,
				},
			},
			want: 150.0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			
			// Arrange
			cart := aggregate.NewCartAggregate()
			for _, existingCmd := range tt.existingItems {
				cart.ExecuteAddItemToCartCommand(existingCmd)
			}

			// Act & Assert
			got := cart.GetTotalAmount()
			assert.Equal(t, tt.want, got.Float64())
		})
	}
}

func TestCartAggregate_Hydration(t *testing.T) {
	cartID := uuid.New()
	userID := uuid.New()
	itemID := uuid.New()

	tests := map[string]struct {
		events      []event.Event
		wantVersion int
	}{
		"should hydrate empty cart with no events": {
			events:      []event.Event{},
			wantVersion: -1,
		},
		"should hydrate cart with CartCreatedEvent": {
			events: []event.Event{
				event.NewCartCreatedEvent(cartID, 1, userID),
			},
			wantVersion: 1,
		},
		"should hydrate cart with full event sequence": {
			events: []event.Event{
				event.NewCartCreatedEvent(cartID, 1, userID),
				event.NewItemAddedToCartEvent(cartID, 2, itemID, 2, 50.0),
				event.NewCartSubmittedEvent(cartID, 3, 100.0),
			},
			wantVersion: 3,
		},
		"should handle adding same item multiple times": {
			events: []event.Event{
				event.NewCartCreatedEvent(cartID, 1, userID),
				event.NewItemAddedToCartEvent(cartID, 2, itemID, 1, 50.0),
				event.NewItemAddedToCartEvent(cartID, 3, itemID, 2, 50.0),
			},
			wantVersion: 3,
		},
		"should handle multiple different items": {
			events: []event.Event{
				event.NewCartCreatedEvent(cartID, 1, userID),
				event.NewItemAddedToCartEvent(cartID, 2, itemID, 1, 50.0),
				event.NewItemAddedToCartEvent(cartID, 3, uuid.New(), 3, 25.0),
			},
			wantVersion: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cart := aggregate.NewCartAggregate()

			// Act & Assert
			err := cart.Hydration(tt.events)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantVersion, cart.GetVersion())
			assert.Len(t, cart.GetUncommittedEvents(), 0)
		})
	}
}

func TestCartAggregate_MarkEventsAsCommitted(t *testing.T) {
	tests := map[string]struct {
		existingItems []command.AddItemToCartCommand
	}{
		"should clear uncommitted events after adding item": {
			existingItems: []command.AddItemToCartCommand{
				{
					CartID:   uuid.New(),
					UserID:   uuid.New(),
					ItemID:   uuid.New(),
					Quantity: 1,
					Price:    100.0,
				},
			},
		},
		"should handle empty events": {
			existingItems: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cart := aggregate.NewCartAggregate()
			for _, existingCmd := range tt.existingItems {
				cart.ExecuteAddItemToCartCommand(existingCmd)
			}

			// Act
			beforeEventCount := len(cart.GetUncommittedEvents())
			cart.MarkEventsAsCommitted()

			// Assert
			assert.Len(t, cart.GetUncommittedEvents(), 0)

			if beforeEventCount > 0 {
				assert.Greater(t, beforeEventCount, 0, "Should have had events before committing")
			}
		})
	}
}
