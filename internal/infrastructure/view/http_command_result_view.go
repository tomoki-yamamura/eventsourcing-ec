package view

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/presenter/viewmodel"
)

type HTTPCommandResultView struct {
	writer http.ResponseWriter
}

func NewHTTPCommandResultView(w http.ResponseWriter) *HTTPCommandResultView {
	return &HTTPCommandResultView{
		writer: w,
	}
}

func (v *HTTPCommandResultView) Render(ctx context.Context, vm *viewmodel.CommandResultViewModel, status int, err error) error {
	v.writer.Header().Set("Content-Type", "application/json")
	v.writer.WriteHeader(status)

	if err != nil {
		errorResponse := map[string]any{
			"status":  "error",
			"message": err.Error(),
		}
		return json.NewEncoder(v.writer).Encode(errorResponse)
	}

	return json.NewEncoder(v.writer).Encode(vm)
}
