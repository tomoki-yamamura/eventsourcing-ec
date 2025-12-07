package deserializer_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/database/eventstore/deserializer"
)

func TestCartCreatedEventDeserializer(t *testing.T) {
	tests := map[string]struct {
		input []byte
		want  *event.CartCreatedEvent
	}{
		"should deserialize valid json": {
			input: []byte(`{
				"AggregateID": "123e4567-e89b-12d3-a456-426614174000",
				"UserID": "123e4567-e89b-12d3-a456-426614174001",
				"TenantID": "123e4567-e89b-12d3-a456-426614174003",
				"EventID": "123e4567-e89b-12d3-a456-426614174002",
				"Timestamp": "2023-01-01T10:00:00Z",
				"Version": 1
			}`),
			want: &event.CartCreatedEvent{
				AggregateID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				UserID:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
				TenantID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174003"),
				EventID:     uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
				Timestamp:   time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
				Version:     1,
			},
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			// Arrange
			deserializer := deserializer.NewCartCreatedEventDeserializer()

			// Act
			got, err := deserializer.Deserialize(tt.input)

			// Assert
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
