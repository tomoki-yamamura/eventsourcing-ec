package router

import (
	"github.com/gorilla/mux"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/query"
)

type Router struct {
	cartAddItemHandler           *command.CartAddItemCommandHandler
	getCartHandler               *query.GetCartQueryHandler
	createTenantPolicyHandler    *command.CreateTenantCartAbandonedPolicyCommandHandler
	updateTenantPolicyHandler    *command.UpdateTenantCartAbandonedPolicyCommandHandler
	getTenantPolicyHandler       *query.GetTenantPolicyQueryHandler
}

func NewRouter(
	cartAddItemHandler *command.CartAddItemCommandHandler,
	getCartHandler *query.GetCartQueryHandler,
	createTenantPolicyHandler *command.CreateTenantCartAbandonedPolicyCommandHandler,
	updateTenantPolicyHandler *command.UpdateTenantCartAbandonedPolicyCommandHandler,
	getTenantPolicyHandler *query.GetTenantPolicyQueryHandler,
) *Router {
	return &Router{
		cartAddItemHandler:        cartAddItemHandler,
		getCartHandler:            getCartHandler,
		createTenantPolicyHandler: createTenantPolicyHandler,
		updateTenantPolicyHandler: updateTenantPolicyHandler,
		getTenantPolicyHandler:    getTenantPolicyHandler,
	}
}

func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Cart routes
	router.HandleFunc("/carts/{aggregate_id}/items", r.cartAddItemHandler.AddItemToCart).Methods("POST")
	router.HandleFunc("/carts/{aggregate_id}", r.getCartHandler.GetCart).Methods("GET")

	// Tenant policy routes
	router.HandleFunc("/tenants/{aggregate_id}/cart-abandoned-policies", r.createTenantPolicyHandler.CreateTenantCartAbandonedPolicy).Methods("POST")
	router.HandleFunc("/tenants/{aggregate_id}/cart-abandoned-policies", r.updateTenantPolicyHandler.UpdateTenantCartAbandonedPolicy).Methods("PUT")
	router.HandleFunc("/tenants/{aggregate_id}/cart-abandoned-policies", r.getTenantPolicyHandler.GetTenantPolicy).Methods("GET")

	return router
}
