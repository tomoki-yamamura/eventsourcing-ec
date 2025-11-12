package view

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter/viewmodel"
)

func TestHTTPTodoListView_Render(t *testing.T) {
	tests := map[string]struct {
		vm             *viewmodel.TodoListVM
		status         int
		err            error
		expectedStatus int
		expectedBody   map[string]any
		checkEmptyBody bool
	}{
		"success case with items": {
			vm: &viewmodel.TodoListVM{
				AggregateID: "test-aggregate-id",
				UserID:      "test-user-id",
				Items: []viewmodel.TodoItem{
					{Text: "Buy groceries"},
					{Text: "Walk the dog"},
				},
				UpdatedAt: "2025-01-01T10:00:00Z",
			},
			status:         http.StatusOK,
			err:            nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"aggregate_id": "test-aggregate-id",
				"user_id":      "test-user-id",
				"updated_at":   "2025-01-01T10:00:00Z",
			},
		},
		"success case with empty items": {
			vm: &viewmodel.TodoListVM{
				AggregateID: "test-aggregate-id",
				UserID:      "test-user-id",
				Items:       []viewmodel.TodoItem{},
				UpdatedAt:   "2025-01-01T10:00:00Z",
			},
			status:         http.StatusOK,
			err:            nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"aggregate_id": "test-aggregate-id",
				"user_id":      "test-user-id",
				"updated_at":   "2025-01-01T10:00:00Z",
			},
		},
		"error case": {
			vm:             nil,
			status:         http.StatusInternalServerError,
			err:            errors.New("todo list not found"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"status":  "error",
				"message": "todo list not found",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			recorder := httptest.NewRecorder()
			view := NewHTTPTodoListView(recorder)

			// Act
			err := view.Render(context.Background(), tt.vm, tt.status, tt.err)
			require.NoError(t, err)

			// Assert
			require.Equal(t, tt.expectedStatus, recorder.Code)
			require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

			if tt.checkEmptyBody {
				require.Equal(t, "", recorder.Body.String())
				return
			}

			var response map[string]any
			err = json.NewDecoder(recorder.Body).Decode(&response)
			require.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				require.Equal(t, expectedValue, response[key])
			}
		})
	}
}
