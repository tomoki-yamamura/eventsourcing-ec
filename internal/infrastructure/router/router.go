package router

import (
	"github.com/gorilla/mux"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/query"
)

type Router struct {
	cartAddItemHandler *command.CartAddItemCommandHandler
	getCartHandler     *query.GetCartQueryHandler
}

func NewRouter(cartAddItemHandler *command.CartAddItemCommandHandler, getCartHandler *query.GetCartQueryHandler) *Router {
	return &Router{
		cartAddItemHandler: cartAddItemHandler,
		getCartHandler:     getCartHandler,
	}
}

func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/additem/{aggregate_id}", r.cartAddItemHandler.AddItemToCart).Methods("POST")
	router.HandleFunc("/carts/{aggregate_id}", r.getCartHandler.GetCart).Methods("GET")

	return router
}
