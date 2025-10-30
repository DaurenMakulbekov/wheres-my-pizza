package posthandler

import (
	"encoding/json"
	"net/http"
	"wheres-my-pizza/tracking-service/internal/core/ports"
)

type HttpHandler struct {
	PostService ports.PostService
}

func NewHandler(postService ports.PostService) *HttpHandler {
	return &HttpHandler{PostService: postService}
}

func (h *HttpHandler) GetOrderByNumber(w http.ResponseWriter, r *http.Request) {

	orderNum := r.PathValue("order_number")
	res, err := h.PostService.GetOrder(orderNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

func (h *HttpHandler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderNum := r.PathValue("order_number")
	res, err := h.PostService.GetOrderStatus(orderNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}
func (h *HttpHandler) GetWorkersStatus(w http.ResponseWriter, r *http.Request) {

	res, err := h.PostService.GetWorkersStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
