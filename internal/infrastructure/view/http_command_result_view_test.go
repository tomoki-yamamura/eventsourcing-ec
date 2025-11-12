package view

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter/viewmodel"
)

func TestHTTPCommandResultView_Render(t *testing.T) {
	tests := map[string]struct {
		vm             *viewmodel.CommandResultViewModel
		status         int
		err            error
		expectedStatus int
		expectedBody   map[string]any
	}{
		"success case": {
			vm: &viewmodel.CommandResultViewModel{
				AggregateID: "test-aggregate-id",
				Version:     1,
				Events: []viewmodel.EventViewModel{
					{
						Type:       "TestEvent",
						Version:    1,
						Data:       map[string]any{"test": "data"},
						OccurredAt: "2025-01-01T10:00:00Z",
					},
				},
				Status:     "success",
				ExecutedAt: "2025-01-01T10:00:00Z",
			},
			status:         http.StatusOK,
			err:            nil,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"aggregateId": "test-aggregate-id",
				"version":     float64(1),
				"status":      "success",
				"executedAt":  "2025-01-01T10:00:00Z",
			},
		},
		"error case": {
			vm:             nil,
			status:         http.StatusBadRequest,
			err:            errors.New("test error message"),
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]any{
				"status":  "error",
				"message": "test error message",
			},
		},
		"error case with nil viewmodel": {
			vm:             nil,
			status:         http.StatusInternalServerError,
			err:            errors.New("internal server error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"status":  "error",
				"message": "internal server error",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			recorder := httptest.NewRecorder()
			view := NewHTTPCommandResultView(recorder)

			// Act
			err := view.Render(context.Background(), tt.vm, tt.status, tt.err)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

			var response map[string]any
			err = json.NewDecoder(recorder.Body).Decode(&response)
			require.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				require.Equal(t, expectedValue, response[key])
			}
		})
	}
}
