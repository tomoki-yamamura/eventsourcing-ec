package service

import (
	"context"

	"github.com/tomoki-yamamura/eventsourcing-ec/internal/domain/event"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/usecase/ports/gateway"
)

type CombinedProjector struct {
	projectors []gateway.Projector
}

func NewCombinedProjector(projectors ...gateway.Projector) gateway.Projector {
	return &CombinedProjector{
		projectors: projectors,
	}
}

func (c *CombinedProjector) Handle(ctx context.Context, e event.Event) error {
	for _, projector := range c.projectors {
		if err := projector.Handle(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

func (c *CombinedProjector) Start(ctx context.Context, bus gateway.EventSubscriber) error {
	for _, projector := range c.projectors {
		if err := projector.Start(ctx, bus); err != nil {
			return err
		}
	}
	return nil
}
