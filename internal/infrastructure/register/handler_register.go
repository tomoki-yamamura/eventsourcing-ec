package register

import (
	"github.com/tomoki-yamamura/eventsourcing-ec/container"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/query"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/router"
)

type HandlerRegister struct {
	container *container.Container
}

func NewHandlerRegister(container *container.Container) *HandlerRegister {
	return &HandlerRegister{
		container: container,
	}
}

func (r *HandlerRegister) SetupRouter() *router.Router {
	// Command handlers
	addItemCommandHandler := command.NewCartAddItemCommandHandler(r.container.CartAddItemCommand)
	createTenantPolicyCommandHandler := command.NewCreateTenantCartAbandonedPolicyCommandHandler(r.container.CreateTenantCartAbandonedPolicyCommand)
	updateTenantPolicyCommandHandler := command.NewUpdateTenantCartAbandonedPolicyCommandHandler(r.container.UpdateTenantCartAbandonedPolicyCommand)

	// Query handlers
	getCartQueryHandler := query.NewGetCartQueryHandler(r.container.GetCartQuery)
	getTenantPolicyQueryHandler := query.NewGetTenantPolicyQueryHandler(r.container.GetTenantPolicyQuery)

	// Router setup
	return router.NewRouter(addItemCommandHandler, getCartQueryHandler, createTenantPolicyCommandHandler, updateTenantPolicyCommandHandler, getTenantPolicyQueryHandler)
}
