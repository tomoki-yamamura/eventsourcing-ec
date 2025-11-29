package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/repository"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/messaging/dto"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
)

type ProjectorService struct {
	transaction   repository.Transaction
	deserializer  repository.EventDeserializer
	cartProjector gateway.Projector
	consumerGroup messaging.ConsumerGroup
}

func NewProjectorService(
	transaction repository.Transaction,
	deserializer repository.EventDeserializer,
	cartProjector gateway.Projector,
	consumerGroup messaging.ConsumerGroup,
) *ProjectorService {
	service := &ProjectorService{
		transaction:   transaction,
		deserializer:  deserializer,
		cartProjector: cartProjector,
		consumerGroup: consumerGroup,
	}

	consumerGroup.AddHandler(service.handleMessage)

	return service
}

func (s *ProjectorService) handleMessage(ctx context.Context, msg *dto.Message) error {
	// Marshal message data for deserializer
	eventData, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	// Deserialize event
	event, err := s.deserializer.Deserialize(msg.Type, eventData)
	if err != nil {
		return err
	}

	// Handle with projector
	return s.cartProjector.Handle(ctx, event)
}

func (s *ProjectorService) Start(ctx context.Context) error {
	log.Println("Starting Projector Service...")
	return s.consumerGroup.Start(ctx)
}

func (s *ProjectorService) Close() error {
	return s.consumerGroup.Close()
}
