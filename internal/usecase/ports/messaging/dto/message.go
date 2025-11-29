package dto

import "github.com/google/uuid"

type Message struct {
	ID          uuid.UUID   `json:"id"`
	Type        string      `json:"type"`
	Data        any `json:"data"`
	AggregateID uuid.UUID   `json:"aggregate_id"`
	Version     int         `json:"version"`
}
