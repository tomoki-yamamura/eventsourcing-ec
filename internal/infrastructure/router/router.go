package router

import (
	"github.com/gorilla/mux"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/command"
	"github.com/tomoki-yamamura/eventsourcing-ec/internal/infrastructure/handler/query"
)

type Router struct {
	createCommandHandler *command.TodoListCreateCommandHandler
	addCommandHandler    *command.TodoAddItemCommandHandler
	queryHandler         *query.TodoListQueryHandler
}

func NewRouter(createCommandHandler *command.TodoListCreateCommandHandler, addCommandHandler *command.TodoAddItemCommandHandler, queryHandler *query.TodoListQueryHandler) *Router {
	return &Router{
		createCommandHandler: createCommandHandler,
		addCommandHandler:    addCommandHandler,
		queryHandler:         queryHandler,
	}
}

func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/todo-lists", r.createCommandHandler.CreateTodoList).Methods("POST")
	router.HandleFunc("/todo-lists/{aggregate_id}/items", r.addCommandHandler.AddTodo).Methods("POST")

	router.HandleFunc("/todo-lists/{aggregate_id}/items", r.queryHandler.Query).Methods("GET")

	return router
}
