package router

import (
	"github.com/gorilla/mux"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/command"
)

type Router struct {
	cartAddItemHandler *command.CartAddItemCommandHandler
}

func NewRouter(cartAddItemHandler *command.CartAddItemCommandHandler) *Router {
	return &Router{
		cartAddItemHandler: cartAddItemHandler,
	}
}

func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/additem/{aggregate_id}", r.cartAddItemHandler.AddItemToCart).Methods("POST")

	return router
}
