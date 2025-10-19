package handlers

import (
	"net/http"
)

func (hd *handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /orders", hd.CreateOrderHandler)

	return mux
}
