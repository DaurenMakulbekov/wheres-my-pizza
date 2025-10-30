package routes

import (
	"net/http"
	"wheres-my-pizza/tracking-service/internal/handlers/posthandler"
)

func SetupRoutes(mux *http.ServeMux, handler *posthandler.HttpHandler) {
	mux.HandleFunc("GET /orders/{order_number}/status", handler.GetOrderByNumber)
	mux.HandleFunc("GET /orders/{order_number}/history", handler.GetOrderStatus)
	mux.HandleFunc("GET /workers/status", handler.GetWorkersStatus)


}
