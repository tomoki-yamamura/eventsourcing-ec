package todo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/projector/todo"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/readmodelstore/dto"
)

type mockViewRepository struct {
	data      map[string]*dto.TodoListViewDTO
	getError  error
	saveError error
}

func (m *mockViewRepository) Get(ctx context.Context, aggregateID string) (*dto.TodoListViewDTO, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	view := m.data[aggregateID]
	if view == nil {
		return nil, errors.NotFound.New("todo list not found")
	}
	return view, nil
}

func (m *mockViewRepository) Upsert(ctx context.Context, aggregateID string, view *dto.TodoListViewDTO) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.data[aggregateID] = view
	return nil
}

func TestTodoProjectorImpl_Handle_TodoListCreatedEvent(t *testing.T) {
	tests := map[string]struct {
		existingView *dto.TodoListViewDTO
		event        event.TodoListCreatedEvent
		want         *dto.TodoListViewDTO
	}{
		"should create new view for TodoListCreatedEvent": {
			existingView: nil,
			event: event.TodoListCreatedEvent{
				AggregateID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				UserID:      mustNewUserID(t, "user123"),
				EventID:     uuid.New(),
				Timestamp:   time.Now(),
				Version:     1,
			},
			want: &dto.TodoListViewDTO{
				AggregateID: "550e8400-e29b-41d4-a716-446655440000",
				UserID:      "user123",
				Items:       []dto.TodoItemViewDTO{},
				Version:     1,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			mockRepo := &mockViewRepository{
				data: make(map[string]*dto.TodoListViewDTO),
			}
			if tt.existingView != nil {
				mockRepo.data[tt.event.AggregateID.String()] = tt.existingView
			}
			projector := todo.NewTodoProjector(mockRepo).(*todo.TodoProjectorImpl)

			// Act
			err := projector.Handle(context.Background(), tt.event)

			// Assert
			require.NoError(t, err)
			saved := mockRepo.data[tt.event.AggregateID.String()]
			require.NotNil(t, saved)
			require.Equal(t, tt.want.AggregateID, saved.AggregateID)
			require.Equal(t, tt.want.UserID, saved.UserID)
			require.Equal(t, tt.want.Version, saved.Version)
			require.Equal(t, len(tt.want.Items), len(saved.Items))
		})
	}
}

func TestTodoProjectorImpl_Handle_TodoAddedEvent(t *testing.T) {
	aggregateID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := map[string]struct {
		existingView *dto.TodoListViewDTO
		event        event.TodoAddedEvent
		want         *dto.TodoListViewDTO
		wantError    error
	}{
		"should add item to existing view": {
			existingView: &dto.TodoListViewDTO{
				AggregateID: aggregateID.String(),
				UserID:      "user123",
				Items:       []dto.TodoItemViewDTO{},
				Version:     1,
			},
			event: event.TodoAddedEvent{
				AggregateID: aggregateID,
				UserID:      mustNewUserID(t, "user123"),
				TodoText:    mustNewTodoText(t, "Buy groceries"),
				EventID:     uuid.New(),
				Timestamp:   time.Now(),
				Version:     2,
			},
			want: &dto.TodoListViewDTO{
				AggregateID: aggregateID.String(),
				UserID:      "user123",
				Items: []dto.TodoItemViewDTO{
					{Text: "Buy groceries"},
				},
				Version: 2,
			},
			wantError: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			mockRepo := &mockViewRepository{
				data: make(map[string]*dto.TodoListViewDTO),
			}
			if tt.existingView != nil {
				mockRepo.data[aggregateID.String()] = tt.existingView
			}
			projector := todo.NewTodoProjector(mockRepo).(*todo.TodoProjectorImpl)

			// Act
			err := projector.Handle(context.Background(), tt.event)

			// Assert
			require.NoError(t, err)
			saved := mockRepo.data[aggregateID.String()]
			require.NotNil(t, saved)
			require.Equal(t, tt.want.AggregateID, saved.AggregateID)
			require.Equal(t, tt.want.UserID, saved.UserID)
			require.Equal(t, tt.want.Version, saved.Version)
			require.Equal(t, len(tt.want.Items), len(saved.Items))
			if len(tt.want.Items) > 0 {
				require.Equal(t, tt.want.Items[0].Text, saved.Items[0].Text)
			}
		})
	}
}

func mustNewUserID(t *testing.T, id string) value.UserID {
	t.Helper()

	userID, err := value.NewUserID(id)
	require.NoError(t, err)
	return userID
}

func mustNewTodoText(t *testing.T, text string) value.TodoText {
	t.Helper()

	todoText, err := value.NewTodoText(text)
	require.NoError(t, err)
	return todoText
}
