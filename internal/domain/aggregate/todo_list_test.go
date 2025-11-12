package aggregate_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/aggregate"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/value"
)

func TestTodoListAggregate_CreateTodoList(t *testing.T) {
	tests := map[string]struct {
		userID          string
		expectedVersion int
		expectedEvents  int
		wantErr         bool
	}{
		"valid user ID": {
			userID:          "user123",
			expectedVersion: 1,
			expectedEvents:  1,
			wantErr:         false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			agg := aggregate.NewTodoListAggregate()
			userID, err := value.NewUserID(tt.userID)
			require.NoError(t, err)
			cmd := command.CreateTodoListCommand{
				UserID: userID,
			}

			// Act
			err = agg.ExecuteCreateTodoListCommand(cmd)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.expectedEvents > 0 {
					require.NotEqual(t, uuid.Nil, agg.GetAggregateID())
					require.Equal(t, userID, agg.GetUserID())
				}
				require.Equal(t, tt.expectedVersion, agg.GetVersion())

				uncommittedEvents := agg.GetUncommittedEvents()
				require.Len(t, uncommittedEvents, tt.expectedEvents)
				if tt.expectedEvents > 0 {
					require.Equal(t, "TodoListCreatedEvent", uncommittedEvents[0].GetEventType())
				}
			}
		})
	}
}

func TestTodoListAggregate_ExecuteAddTodoCommand(t *testing.T) {
	tests := map[string]struct {
		todoText        string
		expectedVersion int
		expectedEvents  int
		wantErr         bool
	}{
		"valid todo text": {
			todoText:        "Learn Event Sourcing",
			expectedVersion: 2,
			expectedEvents:  2,
			wantErr:         false,
		},
		"long todo text": {
			todoText:        "This is a very long todo item that contains a lot of details about what needs to be done",
			expectedVersion: 2,
			expectedEvents:  2,
			wantErr:         false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			userID, err := value.NewUserID("user123")
			require.NoError(t, err)
			agg := aggregate.NewTodoListAggregate()

			createCmd := command.CreateTodoListCommand{
				UserID: userID,
			}
			err = agg.ExecuteCreateTodoListCommand(createCmd)
			require.NoError(t, err)

			todoText, err := value.NewTodoText(tt.todoText)
			require.NoError(t, err)
			addCmd := command.AddTodoCommand{
				AggregateID: agg.GetAggregateID(),
				UserID:      userID,
				TodoText:    todoText,
			}

			// Act
			err = agg.ExecuteAddTodoCommand(addCmd)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedVersion, agg.GetVersion())

				uncommittedEvents := agg.GetUncommittedEvents()
				require.Len(t, uncommittedEvents, tt.expectedEvents)
				require.Equal(t, "TodoListCreatedEvent", uncommittedEvents[0].GetEventType())
				require.Equal(t, "TodoAddedEvent", uncommittedEvents[1].GetEventType())
			}
		})
	}
}

func TestTodoListAggregate_ExecuteAddTodoCommand_ExceedsLimit(t *testing.T) {
	tests := map[string]struct {
		todosToAdd         int
		expectedError      error
		expectedEventCount int
		wantErr            bool
	}{
		"add 4th todo exceeds limit": {
			todosToAdd:         4,
			expectedError:      aggregate.ErrTooManyTodos,
			expectedEventCount: 4,
			wantErr:            true,
		},
		"add exactly 3 todos": {
			todosToAdd:         3,
			expectedError:      nil,
			expectedEventCount: 4,
			wantErr:            false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			userID, err := value.NewUserID("user123")
			require.NoError(t, err)
			agg := aggregate.NewTodoListAggregate()

			// First create the todo list
			createCmd := command.CreateTodoListCommand{
				UserID: userID,
			}
			err = agg.ExecuteCreateTodoListCommand(createCmd)
			require.NoError(t, err)

			var lastErr error
			// Add todos
			for i := range tt.todosToAdd {
				todoText, err := value.NewTodoText(fmt.Sprintf("Todo %d", i+1))
				require.NoError(t, err)
				cmd := command.AddTodoCommand{
					AggregateID: agg.GetAggregateID(),
					UserID:      userID,
					TodoText:    todoText,
				}
				lastErr = agg.ExecuteAddTodoCommand(cmd)
				if lastErr != nil {
					break // Stop on first error
				}
			}

			// Assert
			if tt.wantErr {
				require.Error(t, lastErr)
				require.ErrorIs(t, lastErr, tt.expectedError)
			} else {
				require.NoError(t, lastErr)
			}
			require.Len(t, agg.GetUncommittedEvents(), tt.expectedEventCount)
		})
	}
}
