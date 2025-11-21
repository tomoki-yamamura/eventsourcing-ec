package view

import (
	"encoding/json"
	"net/http"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/errors"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/view"
)

type HTTPQueryResultView struct {
	writer http.ResponseWriter
}

func NewHTTPQueryResultView(writer http.ResponseWriter) view.QueryResultView {
	return &HTTPQueryResultView{
		writer: writer,
	}
}

func (v *HTTPQueryResultView) Success(data []byte) error {
	v.writer.Header().Set("Content-Type", "application/json")
	v.writer.WriteHeader(http.StatusOK)
	_, err := v.writer.Write(data)
	return err
}

func (v *HTTPQueryResultView) Error(err error) error {
	v.writer.Header().Set("Content-Type", "application/json")

	statusCode := v.getStatusCodeFromError(err)
	v.writer.WriteHeader(statusCode)

	errorResponse := map[string]string{
		"error": err.Error(),
	}

	return json.NewEncoder(v.writer).Encode(errorResponse)
}

func (v *HTTPQueryResultView) getStatusCodeFromError(err error) int {
	if errors.IsCode(err, errors.NotFound) {
		return http.StatusNotFound
	}
	if errors.IsCode(err, errors.InvalidParameter) {
		return http.StatusBadRequest
	}
	if errors.IsCode(err, errors.UnpermittedOp) {
		return http.StatusForbidden
	}

	return http.StatusInternalServerError
}
