package presenter

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter/viewmodel"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/query/output"
)

type mockTodoListView struct {
	renderFunc func(ctx context.Context, vm *viewmodel.TodoListVM, status int, err error) error
}

func (m *mockTodoListView) Render(ctx context.Context, vm *viewmodel.TodoListVM, status int, err error) error {
	if m.renderFunc != nil {
		return m.renderFunc(ctx, vm, status, err)
	}
	return nil
}

func TestHTTPTodoListPresenter_Present(t *testing.T) {
	testAggregateID := uuid.New()
	testUserID := uuid.New()
	testTime := time.Now()

	tests := map[string]struct {
		input      *output.GetTodoListOutput
		setupMock  func(*mockTodoListView)
		wantError  error
		wantStatus int
		wantVM     *viewmodel.TodoListVM
	}{
		"successful presentation with items": {
			input: &output.GetTodoListOutput{
				AggregateID: testAggregateID.String(),
				UserID:      testUserID.String(),
				Items: []output.TodoItem{
					{Text: "First todo"},
					{Text: "Second todo"},
				},
				UpdatedAt: testTime,
			},
			setupMock: func(m *mockTodoListView) {
				m.renderFunc = func(ctx context.Context, vm *viewmodel.TodoListVM, status int, err error) error {
					require.Equal(t, http.StatusOK, status)
					require.Nil(t, err)
					require.Equal(t, testAggregateID.String(), vm.AggregateID)
					require.Equal(t, testUserID.String(), vm.UserID)
					require.Len(t, vm.Items, 2)
					require.Equal(t, "First todo", vm.Items[0].Text)
					require.Equal(t, "Second todo", vm.Items[1].Text)
					require.Equal(t, testTime.Format(time.RFC3339), vm.UpdatedAt)
					return nil
				}
			},
			wantError:  nil,
			wantStatus: http.StatusOK,
		},
		"successful presentation with empty items": {
			input: &output.GetTodoListOutput{
				AggregateID: testAggregateID.String(),
				UserID:      testUserID.String(),
				Items:       []output.TodoItem{},
				UpdatedAt:   testTime,
			},
			setupMock: func(m *mockTodoListView) {
				m.renderFunc = func(ctx context.Context, vm *viewmodel.TodoListVM, status int, err error) error {
					require.Equal(t, http.StatusOK, status)
					require.Nil(t, err)
					require.Equal(t, testAggregateID.String(), vm.AggregateID)
					require.Equal(t, testUserID.String(), vm.UserID)
					require.Len(t, vm.Items, 0)
					require.Equal(t, testTime.Format(time.RFC3339), vm.UpdatedAt)
					return nil
				}
			},
			wantError:  nil,
			wantStatus: http.StatusOK,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockView := &mockTodoListView{}
			if tt.setupMock != nil {
				tt.setupMock(mockView)
			}

			presenter := NewHTTPTodoListPresenter(mockView)
			err := presenter.Present(context.Background(), tt.input)

			if tt.wantError != nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantError)
			}
		})
	}
}

func TestHTTPTodoListPresenter_PresentNotFound(t *testing.T) {
	tests := map[string]struct {
		inputError error
		setupMock  func(*mockTodoListView)
	}{
		"not found error presentation": {
			inputError: errors.NotFound.New("todo list not found"),
			setupMock: func(m *mockTodoListView) {
				m.renderFunc = func(ctx context.Context, vm *viewmodel.TodoListVM, status int, err error) error {
					require.Equal(t, http.StatusNotFound, status)
					require.Nil(t, vm)
					require.Error(t, err)
					require.True(t, errors.IsCode(err, errors.NotFound))
					return nil
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockView := &mockTodoListView{}
			if tt.setupMock != nil {
				tt.setupMock(mockView)
			}

			presenter := NewHTTPTodoListPresenter(mockView)
			err := presenter.PresentNotFound(context.Background(), tt.inputError)

			require.NoError(t, err)
		})
	}
}

func TestHTTPTodoListPresenter_PresentError(t *testing.T) {
	tests := map[string]struct {
		inputError error
		setupMock  func(*mockTodoListView)
	}{
		"Error presentation": {
			inputError: errors.Unknown.New("unknow error"),
			setupMock: func(m *mockTodoListView) {
				m.renderFunc = func(ctx context.Context, vm *viewmodel.TodoListVM, status int, err error) error {
					require.Equal(t, http.StatusInternalServerError, status)
					require.Nil(t, vm)
					require.Error(t, err)
					require.True(t, errors.IsCode(err, errors.Unknown))
					return nil
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			mockView := &mockTodoListView{}
			if tt.setupMock != nil {
				tt.setupMock(mockView)
			}

			presenter := NewHTTPTodoListPresenter(mockView)
			err := presenter.PresentError(context.Background(), tt.inputError)

			require.NoError(t, err)
		})
	}
}
